# bitly-lambda - A serverless app to get Bitly statistics

A serverless tool designed to help get statistics from Bitly.

## Layout
```bash
.
├── Makefile                    <-- Makefile to build and deploy
├── event.json                  <-- Sample event to test using SAM local
├── README.md                   <-- This file
├── src                         <-- Source code for a lambda function
│   ├── httpUtil.go             <-- Utils to interact with HTTP requests
│   ├── main.go                 <-- Lambda trigger code
│   ├── s3Util.go               <-- Utils to interact with Amazon S3
│   └── ssmUtil.go              <-- Utils to interact with Amazon SSM
└── template.yaml               <-- SAM Template
```

## Todo
- [ ] Update README.md
- [ ] Update code to store data in a database (like Aurora Serverless)
- [ ] Until we go to a serverless database create a more restrictive policy set for the S3 capabilities