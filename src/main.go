// Package main is the main implementation of the Bitly serverless app.
//
// The app retrieves statistics on the various bitlinks associated with the account of the
// authenticated user
package main

// The imports
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/retgits/bitly-lambda/src/database"
	"github.com/retgits/bitly-lambda/src/util"
)

// Constants
const (
	// The name of the Bitly Generic Access Token parameter in Amazon SSM
	tokenName = "/bitly/apptoken"
	// The name of the database file
	databaseName = "bitly-stats.db"
)

// Variables
var (
	// The region in which the Lambda function is deployed
	awsRegion = util.GetEnvKey("region", "us-west-2")
	// The name of the bucket that has the csv file
	s3Bucket = util.GetEnvKey("s3Bucket", "retgits-bitly")
	// The temp folder to store the database file while working
	tempFolder = util.GetEnvKey("tempFolder", ".")
)

// The handler function is executed every time that a new Lambda event is received.
// It takes a JSON payload (you can see an example in the event.json file) and only
// returns an error if the something went wrong. The event comes fom CloudWatch and
// is scheduled every interval (where the interval is defined as variable)
func handler(request events.CloudWatchEvent) error {
	log.Printf("Processing Lambda request [%s]", request.ID)

	// Create a new session without AWS credentials. This means the Lambda function must have
	// privileges to read and write S3
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))

	// Get the Bitly Generic Access Token
	bitlyToken, err := util.GetSSMParameter(awsSession, tokenName, true)
	if err != nil {
		return err
	}

	// Make a backup of the latest version of the bitly stats
	err = util.CopyFile(awsSession, databaseName, s3Bucket)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Download the latest version of the bitly stats
	err = util.DownloadFile(awsSession, tempFolder, databaseName, s3Bucket)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Get the groups associated with the current account. There should be only one group for a free account
	httpHeader := http.Header{"Authorization": {fmt.Sprintf("Bearer %s", bitlyToken)}}
	response, err := util.HTTPRequest("https://api-ssl.bitly.com/v4/groups", httpHeader)
	if err != nil {
		return err
	}

	bitlyGroup := response["groups"].([]interface{})[0].(map[string]interface{})["guid"].(string)

	// Get the bitlinks
	// TODO: Handle pagination in case more than 50 links are created
	response, err = util.HTTPRequest(fmt.Sprintf("https://api-ssl.bitly.com/v4/groups/%s/bitlinks", bitlyGroup), httpHeader)
	if err != nil {
		return err
	}
	links := response["links"].([]interface{})

	// Get a filehandle to the database
	db, err := database.New(filepath.Join(tempFolder, databaseName))
	if err != nil {
		log.Printf("Error while connecting to the database: %s\n", err.Error())
		return err
	}

	for _, link := range links {
		currLink := link.(map[string]interface{})
		yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

		// Get the link information from Bitly
		response, err = util.HTTPRequest(fmt.Sprintf("https://api-ssl.bitly.com/v4/bitlinks/%s/clicks?unit=day", strings.Replace(currLink["id"].(string), "/", "%2F", 1)), httpHeader)
		if err != nil {
			return err
		}

		// Write the link information to the db
		if len(response["link_clicks"].([]interface{})) > 1 {
			clicks := response["link_clicks"].([]interface{})[1].(map[string]interface{})
			if strings.HasPrefix(clicks["date"].(string), yesterday) {
				u, err := url.Parse(currLink["long_url"].(string))
				if err != nil {
					return err
				}
				m, _ := url.ParseQuery(u.RawQuery)
				item := make(map[string]interface{})
				item["host"] = u.Host
				item["path"] = u.Path
				item["date"] = clicks["date"]
				item["link"] = currLink["link"]
				item["url"] = currLink["long_url"]
				item["clicks"] = clicks["clicks"]
				item["utm_source"] = getValue(m["utm_source"], "")
				item["utm_medium"] = getValue(m["utm_medium"], "")
				item["utm_campaign"] = getValue(m["utm_campaign"], "")
				item["utm_term"] = getValue(m["utm_term"], "")
				item["utm_content"] = getValue(m["utm_content"], "")
				err = db.InsertItem(item)
				if err != nil {
					return err
				}
			}
		}
	}

	util.UploadFile(awsSession, tempFolder, databaseName, s3Bucket)

	return nil
}

func getValue(value []string, fallback string) string {
	if len(value) == 0 {
		return fallback
	}

	return value[0]
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
