/*
Package main is the main implementation of the Bitly serverless app and retrieves statistics on the various bitlinks associated with the account of the authenticated user
*/
package main

// The imports
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Variables
var (
	// The region in which the Lambda function is deployed
	awsRegion = os.Getenv("region")
	// The name of the bucket that has the csv file
	s3Bucket = os.Getenv("s3bucket")
)

// Constants
const (
	// The name of the Bitly Generic Access Token parameter in Amazon SSM
	tokenName = "/bitly/apptoken"
	// The temp folder in AWS Lambda
	tempFolder = "/tmp"
	// The name of the csv file
	csv = "bitly-stats.csv"
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
	bitlyToken, err := getSSMParameter(awsSession, tokenName, true)
	if err != nil {
		return err
	}

	// Get the groups associated with the current account. There should be only one group for a free account
	response, err := httpRequest("https://api-ssl.bitly.com/v4/groups", bitlyToken)
	if err != nil {
		return err
	}
	bitlyGroup := response["groups"].([]interface{})[0].(map[string]interface{})["guid"].(string)

	// Get the bitlinks
	// TODO: Handle pagination in case more than 50 links are created
	response, err = httpRequest(fmt.Sprintf("https://api-ssl.bitly.com/v4/groups/%s/bitlinks", bitlyGroup), bitlyToken)
	if err != nil {
		return err
	}
	links := response["links"].([]interface{})

	// Make a backup of the latest version of the bitly stats
	// TODO: Remove this as soon as data will get stored in a database (like Aurora Serverless)
	err = copyFile(awsSession, csv, s3Bucket)
	if err != nil {
		return err
	}

	// Download the latest version of the bitly stats
	// TODO: Remove this as soon as data will get stored in a database (like Aurora Serverless)
	err = getFile(awsSession, csv, s3Bucket)
	if err != nil {
		return err
	}

	// Open a file handle
	// TODO: Remove this as soon as data will get stored in a database (like Aurora Serverless)
	csvFile, err := os.OpenFile(filepath.Join(tempFolder, csv), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Get the clickdata
	for _, link := range links {
		link := link.(map[string]interface{})
		response, err = httpRequest(fmt.Sprintf("https://api-ssl.bitly.com/v4/bitlinks/%s/clicks?unit=day", strings.Replace(link["id"].(string), "/", "%2F", 1)), bitlyToken)
		if err != nil {
			return err
		}
		clicks := response["link_clicks"].([]interface{})[1].(map[string]interface{})
		csvFile.WriteString(fmt.Sprintf("%s;%s;%s;%v", link["link"], link["long_url"], clicks["date"], clicks["clicks"]))
	}

	// Store the modified csv on S3
	// TODO: Remove this as soon as data will get stored in a database (like Aurora Serverless)
	err = uploadFile(awsSession, csv, s3Bucket)
	if err != nil {
		return err
	}

	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
