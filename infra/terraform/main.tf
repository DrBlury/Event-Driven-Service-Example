terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  endpoints {
    sns = "http://localstack:4566"
    sqs = "http://localstack:4566"
    sts = "http://localstack:4566"
  }
}

data "aws_caller_identity" "current" {}
output "is_localstack" {
  value = data.aws_caller_identity.current.id == "000000000000"
}

locals {
  queues = [
    "messages",
    "messages-processed",
    "messages-poison",
    "signup",
    "signup-processable",
  ]

  topics = [
    "messages",
    "messages-processed",
    "messages-poison",
    "signup",
    "signup-processable",
  ]
}

# Create SQS queues
resource "aws_sqs_queue" "queues" {
  for_each = toset(local.queues)
  name     = each.key
}

# Create SNS topics
resource "aws_sns_topic" "topics" {
  for_each = toset(local.topics)
  name     = each.key
}

# Subscribe each topic to the same-named SQS queue
resource "aws_sns_topic_subscription" "topic_subs" {
  for_each = {
    for t in local.topics : t => t
  }

  topic_arn = aws_sns_topic.topics[each.key].arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.queues[each.key].arn
}

# Allow SNS to send messages to SQS queues (queue policy)
resource "aws_sqs_queue_policy" "allow_sns" {
  for_each = aws_sqs_queue.queues

  queue_url = each.value.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "Allow-SNS-SendMessage"
        Effect    = "Allow"
        Principal = "*"
        Action    = "sqs:SendMessage"
        Resource  = each.value.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = [for t in aws_sns_topic.topics : t.arn]
          }
        }
      }
    ]
  })
}
