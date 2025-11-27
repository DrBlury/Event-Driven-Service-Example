variable "queues" {
  type = map(string)
}

variable "topic_arns" {
  type = list(string)
}

variable "sqs_managed_sse_enabled" {
  type        = bool
  default     = true
  description = "Enable SQS managed server-side encryption"
}

resource "aws_sqs_queue" "queues" {
  for_each                = var.queues
  name                    = each.key
  sqs_managed_sse_enabled = var.sqs_managed_sse_enabled # Enable server-side encryption
}

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
            "aws:SourceArn" = var.topic_arns
          }
        }
      }
    ]
  })
}
