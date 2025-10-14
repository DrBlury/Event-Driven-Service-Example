#!/bin/bash

export AWS_ACCESS_KEY_ID=000000000000 AWS_SECRET_ACCESS_KEY=000000000000

# create SNS topics
awslocal sns create-topic --name messages
awslocal sns create-topic --name messages-processed
awslocal sns create-topic --name messages-poison
awslocal sns create-topic --name signup
awslocal sns create-topic --name signup-processeable

# create subscriptions to SQS queues (assuming the queues are already created)
awslocal sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:messages --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:messages
awslocal sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:messages-processed --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:messages-processed
awslocal sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:messages-poison --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:messages-poison
awslocal sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:signup --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:signup
awslocal sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:signup-processable --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:signup-processable
