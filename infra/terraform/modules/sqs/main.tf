variable "queues" {
  type = map(string)
}

variable "topic_arns" {
  type = list(string)
}

resource "aws_sqs_queue" "queues" {
  for_each = var.queues
  name     = each.key
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
