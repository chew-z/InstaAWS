# InstaAWS

AWS Lambda function in Go for getting some metadata about user Instagram feed images. The goal is to use Instagram as free CDN for some hugo websites, passing most of traffic bills to Zuckenberg. My webpage itself is c.a. 40 kB.

## Status

Works well. Requires some tuning and further development.

## How to use it?

- set ENV variables (username, password) in .env.yml
- deploy using serverless as AWS Lambda - Go function [see serverless.yml]
```
   sls deploy 
```
- get [saw](https://github.com/donnemartin/saws) and watch logs
```
  saw watch /aws/lambda/endpoint --region "eu-central-1"
```
- get JSON file with pictures metatdata from user feed
``` 
    curl -X GET -o instagram.json 'uri://lambda-ednpoint?limit=100' 
```

## TODO

- get metadata in chunks
- get random pictures

## Limitations

AWS [API Gateway has intergration timeout](https://docs.aws.amazon.com/apigateway/latest/developerguide/limits.html) of maximum 30 seconds. With large Instagram feed and extensive use of append operation in code this isn't enough time for AWS Lambda function to finish. So I am thinking about workarounds like getting data in chunks, narrowing via tags or timestamps or search conditions etc.  But that makes simple code into complex one.

## What's interesting about it?

- I am using [goinsta](https://github.com/ahmdrz/goinsta) - good but it has it's limitations as Import/Export operations are hardcoded for storing session in file - not possible with Lambda. So my login() function is storing Instagram object in cache instead.
- I am using [go-cache](https://github.com/patrickmn/go-cache) to good effect. Now [Lambda functions](https://aws.amazon.com/lambda/faqs/#) are essentially stateless and not reusable (lifetime is circa 60 seconds apparently). However it seems that a series of calls to function can find it's way to function instance and re-use Instagram object from cache. It shaves quite some time on successive calls - watch it happen in logs - I guess it is  taking advantage of [freeze/thaw cycle](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/).
