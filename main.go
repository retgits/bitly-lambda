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
	"github.com/retgits/bitly-lambda/database"
	"github.com/retgits/bitly-lambda/util"
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

	// Prepare a channel for done
	done := make(chan struct{})
	linksChan := make(chan map[string]interface{})

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

	// Create a connection to the SQLite database
	dbase, err := database.New(filepath.Join(tempFolder, databaseName))
	if err != nil {
		log.Printf("Error while connecting to the database: %s\n", err.Error())
		return err
	}

	// Get the groups associated with the current account. There should be only one group for a free account
	httpHeader := http.Header{"Authorization": {fmt.Sprintf("Bearer %s", bitlyToken)}}
	response, err := util.HTTPRequest("https://api-ssl.bitly.com/v4/groups", httpHeader)
	if err != nil {
		log.Printf("Error while connecting to Bitly: %s\n", err.Error())
		return err
	}

	bitlyGroup := response.Body["groups"].([]interface{})[0].(map[string]interface{})["guid"].(string)

	// Get the bitlinks
	// TODO: Handle pagination in case more than 50 links are created
	response, err = util.HTTPRequest(fmt.Sprintf("https://api-ssl.bitly.com/v4/groups/%s/bitlinks", bitlyGroup), httpHeader)
	if err != nil {
		log.Printf("Error while retrieving bitlinks: %s\n", err.Error())
		return err
	}
	links := response.Body["links"].([]interface{})

	// Send each link to a channel
	go func() {
		for _, link := range links {
			linksChan <- link.(map[string]interface{})
		}
	}()

	// Handle the links and throw an error if anything goes wrong
	go func() error {
		// Keep a counter for all of the links. After all links have been handled the program is done
		counter := len(links)
		// The program runs each night and should only gather stats from the day before if there are any
		yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
		for {
			// Receive the link
			link := <-linksChan

			// Get the link information from Bitly
			response, err = util.HTTPRequest(fmt.Sprintf("https://api-ssl.bitly.com/v4/bitlinks/%s/clicks?unit=day", strings.Replace(link["id"].(string), "/", "%2F", 1)), httpHeader)
			if err != nil {
				log.Printf("Error while retrieving bitlink click details: %s\n", err.Error())
				return err
			}

			// Write the link information to the db
			if len(response.Body["link_clicks"].([]interface{})) > 1 {
				clicks := response.Body["link_clicks"].([]interface{})[1].(map[string]interface{})
				if strings.HasPrefix(clicks["date"].(string), yesterday) {
					u, err := url.Parse(link["long_url"].(string))
					if err != nil {
						log.Printf("Error while parsing long_url: %s\n", err.Error())
						return err
					}
					queryParams, _ := url.ParseQuery(u.RawQuery)
					err = dbase.Exec(fmt.Sprintf("insert into links (host, path, date, link, url, clicks, utm_source, utm_medium, utm_campaign, utm_term, utm_content) values (\"%s\",\"%s\",\"%v\",\"%s\",\"%s\",\"%d\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\");", u.Host, u.Path, clicks["date"], link["link"], link["long_url"], clicks["clicks"], getValue(queryParams["utm_source"], ""), getValue(queryParams["utm_medium"], ""), getValue(queryParams["utm_campaign"], ""), getValue(queryParams["utm_term"], ""), getValue(queryParams["utm_content"], "")))
					if err != nil {
						log.Printf("Error while inserting data into the database: %s\n", err.Error())
						return err
					}
				}
			}

			// Decrement the counter
			if counter--; counter == 0 {
				close(done)
			}
		}
	}()

	// Wait for all links to be processed
	<-done

	// Upload the latest version of the SQLite database to S3
	util.UploadFile(awsSession, tempFolder, databaseName, s3Bucket)
	dbase.Close()

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
