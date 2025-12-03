# Terraform Modules

Reusable Terraform modules for infrastructure provisioning.

## Available Modules

### IAM

**Path:** `iam/`

Creates IAM roles and policies for AWS services.

### SNS

**Path:** `sns/`

Provisions Simple Notification Service (SNS) topics with access policies.

### SQS

**Path:** `sqs/`

Creates Simple Queue Service (SQS) queues with optional dead-letter queues and SNS subscriptions.

## Usage Example

See `../environments/localstack/` for a working example:

```hcl
module "sns" {
  source      = "../../modules/sns"
  topic_name  = "example-events"
  environment = "dev"
}

module "sqs" {
  source     = "../../modules/sqs"
  queue_name = "example-queue"
  # ... additional configuration
}

```

## Adapting for Production

**Centralized Module Repository:**

```hcl
module "sns" {
  source = "git::https://github.com/org/terraform-modules.git//sns?ref=v1.0.0"
  # ... configuration
}
```

**Terraform Registry:**

```hcl
module "sns" {
  source  = "app.terraform.io/org/sns/aws"
  version = "~> 1.0"
  # ... configuration
}
```

## Local Testing

Initialize and test locally with LocalStack:

```bash
cd ../environments/localstack
terraform init
terraform plan
terraform apply
```

## Documentation

- [Infrastructure Guide](../../../docs/infrastructure.md) - Comprehensive infrastructure documentation
- [Terraform README](../README.md) - Getting started with Terraform in this project
