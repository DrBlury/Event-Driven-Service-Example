# Terraform Infrastructure

This directory contains Terraform configurations for provisioning cloud resources.

## Quick Start (LocalStack)

The LocalStack environment demonstrates infrastructure patterns locally:

```bash
# Via Docker Compose (automated)
task up-aws

# Manual execution
cd environments/localstack
terraform init
terraform plan
terraform apply
```

## Structure

```
terraform/
├── modules/              # Reusable Terraform modules
│   ├── iam/             # IAM roles and policies
│   ├── sns/             # SNS topics
│   └── sqs/             # SQS queues and subscriptions
└── environments/         # Environment-specific configs
    └── localstack/      # LocalStack demo
```

## Modules

### IAM Module
Creates IAM roles and policies for AWS services.

**Path:** `modules/iam/`

### SNS Module
Provisions SNS topics with configurable access policies.

**Path:** `modules/sns/`

### SQS Module
Creates SQS queues with dead-letter queue support and SNS subscriptions.

**Path:** `modules/sqs/`

## LocalStack Environment

**What it provisions:**
- SNS topics for event publishing
- SQS queues for event consumption
- Queue subscriptions to topics
- IAM policies for service access

**Resources created:**
- SNS: `example-events-topic`
- SQS: `example-events-queue`, `example-events-dlq`
- Subscriptions: SQS → SNS

**Access:**
- LocalStack endpoint: http://localhost:4566
- Health check: http://localhost:4566/_localstack/health

## Adapting for Production

### 1. Move to Central Repository

For production, move modules to a shared infrastructure repository:

```hcl
module "sns" {
  source = "git::https://github.com/org/terraform-modules.git//sns?ref=v1.0.0"
  # ... configuration
}
```

### 2. Create Environment Configs

```
environments/
├── dev/
├── staging/
└── production/
```

### 3. Use Remote State

```hcl
terraform {
  backend "s3" {
    bucket = "company-terraform-state"
    key    = "service/production/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### 4. Manage Secrets Externally

- AWS Secrets Manager
- HashiCorp Vault
- AWS Systems Manager Parameter Store

**Never commit credentials or secrets!**

## Prerequisites

**For LocalStack (via Docker):**
- Docker Compose
- `task up-aws` handles everything

**For manual Terraform:**
```bash
# Install Terraform v1.5+
brew install terraform  # macOS
# or download from https://www.terraform.io/downloads

# Set AWS credentials (LocalStack accepts any values)
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1
export AWS_ENDPOINT=http://localhost:4566  # For LocalStack
```

## Common Commands

```bash
# Initialize
terraform init

# Plan changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy

# Show current state
terraform show

# List resources
terraform state list
```

## Detailed Documentation

See [Infrastructure Guide](../../docs/infrastructure.md) for:
- Complete Terraform workflow
- Module usage examples
- Production deployment patterns
- Security best practices
- Troubleshooting

