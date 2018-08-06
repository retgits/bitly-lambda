.PHONY: deps clean build deploy test-lambda

deps:
	go get -u ./...

clean: 
	rm -rf ./bin
	
build:
	GOOS=linux GOARCH=amd64 go build -o ./bin/bitly-lambda src/*.go

test-lambda: clean build
	sam local invoke bitly -e event.json

deploy:
	sam package --template-file template.yaml --output-template-file packaged.yaml --s3-bucket retgits-bitly
	sam deploy --template-file packaged.yaml --stack-name bitly-lambda --capabilities CAPABILITY_IAM