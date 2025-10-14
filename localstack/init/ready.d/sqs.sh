#!/bin/bash

export AWS_ACCESS_KEY_ID=000000000000 AWS_SECRET_ACCESS_KEY=000000000000

# create SQS queues
awslocal sqs create-queue --queue-name messages
awslocal sqs create-queue --queue-name messages-processed
awslocal sqs create-queue --queue-name messages-poison
awslocal sqs create-queue --queue-name signup
awslocal sqs create-queue --queue-name signup-processable
