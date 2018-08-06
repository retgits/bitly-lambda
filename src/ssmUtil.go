/*
Package main is the main implementation of the Bitly serverless app and retrieves statistics on the various bitlinks associated with the account of the authenticated user
*/
package main

// The imports
import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// getSSMParameter gets a parameter from the AWS Simple Systems Manager service.
func getSSMParameter(awsSession *session.Session, name string, decrypt bool) (string, error) {
	// Create an instance of the SSM Session
	ssmSession := ssm.New(awsSession)

	// Create the request to SSM
	getParameterInput := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(decrypt),
	}

	// Get the parameter from SSM
	param, err := ssmSession.GetParameter(getParameterInput)
	if err != nil {
		return "", err
	}

	return *param.Parameter.Value, nil
}
