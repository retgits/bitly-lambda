# bitly-lambda - A serverless app to get Bitly statistics

A serverless tool designed to help get statistics from Bitly.

## Layout
```bash
.                    
├── database            
│   └── database.go     <-- Utils to interact with the SQLite database
├── test            
│   └── event.json      <-- Sample event to test using SAM local
├── util                
│   ├── http.go         <-- Utils to interact with HTTP requests
│   ├── os.go           <-- Utils to interact with the Operating System
│   ├── s3.go           <-- Utils to interact with Amazon S3
│   └── ssm.go          <-- Utils to interact with Amazon SSM
├── .gitignore          <-- Ignoring the things you do not want in git
├── LICENSE             <-- The license file
├── main_test.go        <-- Test the code
├── main.go             <-- The Lambda code
├── Makefile            <-- Makefile to build and deploy
├── README.md           <-- This file
└── template.yaml       <-- SAM Template
```

## Installing
There are a few ways to install this project

### Get the sources
You can get the sources for this project by simply running
```bash
$ go get -u github.com/retgits/bitly-lambda/...
```

### Deploy
Deploy the Lambda app by running
```bash
$ make deploy
```

## Parameters
### AWS Systems Manager parameters
The code will automatically retrieve the below list of parameters from the AWS Systems Manager Parameter store:

* **/bitly/apptoken**: Your personal OAuth token (check the Bitly documentation on how to get this parameter)

### Deployment parameters
In the `template.yaml` there are certain deployment parameters:

* **region**: The AWS region in which the code is deployed
* **s3Bucket**: The name of the S3 bucket in which the SQLite file is stored
* **tempFolder**: The temp folder where to store the SQLite file while processing (for deployments to AWS Lambda, this should be set to `/tmp`)

## Make targets
biylt-lambda has a _Makefile_ that can be used for most of the operations

```
usage: make [target]
```

* **deps**: Gets all dependencies for this app
* **clean** : Removes the dist directory
* **build**: Builds an executable to be deployed to AWS Lambda
* **test-lambda**: Clean, builds and tests the code by using the AWS SAM CLI
* **deploy**: Cleans, builds and deploys the code to AWS Lambda