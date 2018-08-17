# InstaAWS

AWS Lambda function in Go for getting some metadata about user Instagram feed images. The goal is to use Instagram as free CDN for some hugo websites, passing most of traffic bills to Zuckenberg. My webpage itself is c.a. 40 kB.

## Status

Works well. Requires some tuning and further development.

## How to use it?

- set ENV variables (username, password) in .env.yml
- deploy using serverless as AWS Lambda - Go function [see serverless.yml]
```
   sls deploy ```
- get JSON file with pictures metatdata from user feed
``` 
    curl -X GET -o instagram.json 'lambda-uri?limit=100' ```

## TODO

- get metadata in chunks
- get random pictures

## Limitations

AWS Proxy has maximum timeout of 30 seconds. With large feed and extensive use of append operation in code this isn't enough for function to finish. So I am thinking about workarounds like getting data in chunks, narrowing via tags or timestamps or search conditions etc.  But that makes simple code into complex one.
