terraform {
  # Require a modern Terraform CLI to ensure compatibility with the
  # AWS provider plugin framework used by recent provider releases.
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      # Use the current v6.x AWS provider series (latest tested in Oct 2025).
      # This will be resolved to a concrete version by `terraform init` and
      # recorded in .terraform.lock.hcl for reproducible installs.
      version = "~> 6.0"
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
  queues = {
    messages           = "messages"
    messages_processed = "messages-processed"
    messages_poison    = "messages-poison"
    some               = "some"
    some_processed     = "some-processed"
  }

  topics = {
    messages           = "messages"
    messages_processed = "messages-processed"
    messages_poison    = "messages-poison"
    some               = "some"
    some_processed     = "some-processed"
  }
}

module "sns" {
  source            = "../../modules/sns"
  topics            = local.topics
  kms_master_key_id = aws_kms_key.sns_key.arn
}

module "sqs" {
  source     = "../../modules/sqs"
  queues     = local.queues
  topic_arns = [for t in module.sns : t.arn]
}

module "iam" {
  source     = "../../modules/iam"
  topic_arns = [for t in module.sns : t.arn]
  queue_arns = [for q in module.sqs : q.arn]
}

# Create a customer managed KMS key for SNS encryption
resource "aws_kms_key" "sns_key" {
  description             = "Customer managed KMS key for SNS topic encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow SNS to use the key"
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey*"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_kms_alias" "sns_key_alias" {
  name          = "alias/sns-cmk"
  target_key_id = aws_kms_key.sns_key.key_id
}

# Create SQS queues
resource "aws_sqs_queue" "queues" {
  for_each                = local.queues
  name                    = each.value
  sqs_managed_sse_enabled = true # Enable server-side encryption
}

# Create SNS topics
resource "aws_sns_topic" "topics" {
  for_each          = local.topics
  name              = each.value
  kms_master_key_id = aws_kms_key.sns_key.arn # Enable server-side encryption with customer managed key
}

# Subscribe each topic to the same-named SQS queue
resource "aws_sns_topic_subscription" "topic_subs" {
  for_each = local.topics

  topic_arn = aws_sns_topic.topics[each.key].arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.queues[each.key].arn
}
