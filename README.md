# bitly-lambda - A serverless app to get Bitly statistics

A serverless tool designed to help get statistics from Bitly.

## Layout
```bash
.
├── src                     
│   ├── database            
│   │   └── database.go     <-- Utils to interact with the SQLite database
│   ├── util                
│   │   ├── http.go         <-- Utils to interact with HTTP requests
│   │   ├── os.go           <-- Utils to interact with the Operating System
│   │   ├── s3.go           <-- Utils to interact with Amazn S3
│   │   └── ssm.go          <-- Utils to interact with Amazon SSM
│   ├── main_test.go        <-- A test function
│   └── main.go             <-- Lambda trigger code
├── .gitignore              <-- Ignoring the things you don't want in git
├── event.json              <-- Sample event to test using SAM local
├── LICENSE                 <-- The license file
├── Makefile                <-- Makefile to build and deploy
├── README.md               <-- This file
└── template.yaml           <-- SAM Template
```

## Todo
- [ ] Update README.md
