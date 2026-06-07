terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Remote state backend- commented out for the case study.
  # Uncomment and configure before running `terraform apply` in a real environment.
  #
  # Standard production pattern: S3 bucket for state, DynamoDB table for locking.
  # The locking table prevents concurrent applies from corrupting the state file.
  #
  # backend "s3" {
  #   bucket         = "csv-processor-terraform-state"
  #   key            = "csv-processor/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "csv-processor-terraform-locks"
  # }
}

provider "aws" {
  region = var.aws_region

  # Credentials are never hardcoded. The provider resolves them in order:
  #   1. Environment variables  (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY)
  #   2. ~/.aws/credentials     (local development)
  #   3. IRSA / instance profile (when running in Kubernetes or EC2)
}
