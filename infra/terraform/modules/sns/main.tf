variable "topics" {
  type = map(string)
}

variable "kms_master_key_id" {
  type        = string
  description = "KMS key ARN for SNS encryption. Must be a customer managed key."
  # No default - caller must provide a CMK ARN to satisfy security requirements
}

resource "aws_sns_topic" "topics" {
  for_each          = var.topics
  name              = each.key
  kms_master_key_id = var.kms_master_key_id # Enable server-side encryption
}
