variable "topic_arns" {
  type = list(string)
}

variable "queue_arns" {
  type = list(string)
}

resource "aws_sns_topic_subscription" "topic_subs" {
  for_each = toset(var.topic_arns)

  topic_arn = each.key
  protocol  = "sqs"
  endpoint  = var.queue_arns[0] # Adjust logic if multiple queues
}
