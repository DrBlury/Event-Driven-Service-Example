The modules directory contains reusable Terraform modules that can be utilized across different infrastructure deployments. Each module is designed to encapsulate specific functionality, making it easier to manage and maintain infrastructure as code.

Please do a terraform init in the environments/localstack directory to ensure that the modules are properly referenced and initialized.

## Available Modules
- **IAM**: IAM roles and policies for networking components.
- **SNS**: Simple Notification Service topics and subscriptions.
- **SQS**: Simple Queue Service queues and related configurations.

## Usage
Check the localstack implementation in the `deploy/infra/terraform/environments/localstack` directory for examples of how to use these modules in your Terraform configurations.

For local environment setups, you can directly reference these modules in your Terraform files located in the respective environment directories.

For central infrastructure repo, move the modules to a shared location and reference them accordingly in your Terraform code. Then add the environment-specific configurations in the central infra repo.
