variable "topics" {
  type = map(string)
}

variable "kms_master_key_id" {
  type        = string
  default     = "alias/aws/sns"
  description = "KMS key for SNS encryption. Defaults to AWS managed key."
}

resource "aws_sns_topic" "topics" {
  for_each          = var.topics
  name              = each.key
  kms_master_key_id = var.kms_master_key_id # Enable server-side encryption
}
