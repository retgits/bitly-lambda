// Package main is helper...
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/retgits/lambda-util"
)

// Constants
const (
	// The name of the Bitly Generic Access Token parameter in Amazon SSM
	tokenName = "/bitly/apptoken"
)

// Variables
var (
	// The region in which the Lambda function is deployed
	awsRegion = util.GetEnvKey("region", "us-west-2")
)

func main() {
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
		panic(err)
	}

	// Get the groups associated with the current account. There should be only one group for a free account
	httpHeader := http.Header{"Authorization": {fmt.Sprintf("Bearer %s", bitlyToken)}}
	response, err := util.HTTPGet("https://api-ssl.bitly.com/v4/groups", "application/json", httpHeader)
	if err != nil {
		log.Printf("Error while connecting to Bitly: %s\n", err.Error())
		panic(err)
	}

	bitlyGroup := response.Body["groups"].([]interface{})[0].(map[string]interface{})["guid"].(string)

	// Get the bitlinks
	// TODO: Handle pagination in case more than 50 links are created
	response, err = util.HTTPGet(fmt.Sprintf("https://api-ssl.bitly.com/v4/groups/%s/bitlinks", bitlyGroup), "application/json", httpHeader)
	if err != nil {
		log.Printf("Error while retrieving bitlinks: %s\n", err.Error())
		panic(err)
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
			response, err = util.HTTPGet(fmt.Sprintf("https://api-ssl.bitly.com/v4/bitlinks/%s/clicks?unit=day", strings.Replace(link["id"].(string), "/", "%2F", 1)), "application/json", httpHeader)
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
					fmt.Printf("insert into links (host, path, date, link, url, clicks, utm_source, utm_medium, utm_campaign, utm_term, utm_content) values (\"%s\",\"%s\",\"%v\",\"%s\",\"%s\",\"%d\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\");\n", u.Host, u.Path, clicks["date"], link["link"], link["long_url"], clicks["clicks"], getValue(queryParams["utm_source"], ""), getValue(queryParams["utm_medium"], ""), getValue(queryParams["utm_campaign"], ""), getValue(queryParams["utm_term"], ""), getValue(queryParams["utm_content"], ""))
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
}

func getValue(value []string, fallback string) string {
	if len(value) == 0 {
		return fallback
	}

	return value[0]
}
