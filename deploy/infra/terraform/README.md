This folder contains Terraform configuration to provision SNS topics, SQS queues and the subscriptions between them for LocalStack.

Usage (via Docker Compose in this repo): the `terraform` helper service will run `terraform init` and `terraform apply` against the LocalStack endpoint.

If you want to run locally:

1. Install Terraform v1.5+.
2. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY (LocalStack accepts anything).
3. terraform init && terraform apply -auto-approve
