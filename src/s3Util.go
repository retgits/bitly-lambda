/*
Package main is the main implementation of the Bitly serverless app and retrieves statistics on the various bitlinks associated with the account of the authenticated user
*/
package main

// The imports
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// getFile downloads a specific file to a temporary location in AWS Lambda
func getFile(awsSession *session.Session, filename string, s3Bucket string) error {
	// Create an instance of the S3 Downloader
	s3Downloader := s3manager.NewDownloader(awsSession)

	// Create a new temporary file
	tempFile, err := os.Create(filepath.Join(tempFolder, filename))
	if err != nil {
		return err
	}

	// Prepare the download
	objectInput := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(filename),
	}

	// Download the file to disk
	_, err = s3Downloader.Download(tempFile, objectInput)
	if err != nil {
		return err
	}

	return nil
}

// copyFile creates a copy of an existing file with a new name
func copyFile(awsSession *session.Session, filename string, s3Bucket string) error {
	// Create an instance of the S3 Session
	s3Session := s3.New(awsSession)

	// Prepare the copy object
	objectInput := &s3.CopyObjectInput{
		Bucket:     aws.String(s3Bucket),
		CopySource: aws.String(fmt.Sprintf("/%s/%s", s3Bucket, filename)),
		Key:        aws.String(fmt.Sprintf("%s_bak", filename)),
	}

	// Copy the object
	_, err := s3Session.CopyObject(objectInput)
	if err != nil {
		return err
	}

	return nil
}

// uploadFile uploads a file to S3
func uploadFile(awsSession *session.Session, filename string, s3Bucket string) error {
	// Create an instance of the S3 Uploader
	s3Uploader := s3manager.NewUploader(awsSession)

	// Create a file pointer to the source
	reader, err := os.Open(filepath.Join(tempFolder, filename))
	if err != nil {
		return err
	}
	defer reader.Close()

	// Prepare the upload
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(filename),
		Body:   reader,
	}

	// Upload the file
	_, err = s3Uploader.Upload(uploadInput)
	if err != nil {
		return err
	}

	return nil
}
