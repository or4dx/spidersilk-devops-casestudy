variable "env" {
  description = "Deployment environment. Appended to the bucket name: <prefix>-<env>."
  type        = string
  default     = "dev"

  # For a multi-environment setup, override at apply time:
  #   terraform apply -var="env=staging"
  #   terraform apply -var="env=prod"
  # This produces: csv-processor-uploads-staging, csv-processor-uploads-prod.
  # Each environment gets its own isolated bucket and lifecycle policy.
}

variable "aws_region" {
  description = "AWS region to deploy into."
  type        = string
  default     = "us-east-1"
  # us-east-1: deepest spot market, broadest AWS service availability.
  # Matches the KOPS cluster region in k8s-cluster/cluster.yaml.
}

variable "bucket_name_prefix" {
  description = "Prefix for the S3 bucket name. Final name: <prefix>-<env>."
  type        = string
  default     = "csv-processor-uploads"
}
