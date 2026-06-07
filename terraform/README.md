# Terraform — S3 Bucket

This configuration provisions the S3 bucket used by the CSV processor app to
store uploaded files under the `processed/` prefix.

## Status

**Configuration complete. Not applied.**

The remote state backend is commented out in `main.tf`. Before running
`terraform apply`, you must:

1. Create an S3 bucket and DynamoDB table for remote state (or use local state
   for a quick test).
2. Uncomment and configure the `backend "s3"` block in `main.tf`.
3. Run `terraform init` to initialise the backend.
4. Ensure AWS credentials are available (env vars, `~/.aws/credentials`, or
   an EC2/EKS instance profile).

```bash
cd terraform/
terraform init
terraform plan     # validate — no changes applied
terraform apply    # provisions the bucket
```

## What is created

| Resource | Description |
|---|---|
| `aws_s3_bucket` | `csv-processor-uploads-dev` (default env) |
| `aws_s3_bucket_versioning` | Versioning enabled |
| `aws_s3_bucket_server_side_encryption_configuration` | AES256 at rest |
| `aws_s3_bucket_public_access_block` | All public access blocked |
| `aws_s3_bucket_lifecycle_configuration` | Tiered archival on `processed/` prefix |

## Outputs

After `terraform apply`, the outputs connect Terraform to the rest of the stack:

```bash
terraform output bucket_name   # → use in ansible/group_vars/all.yaml: s3_bucket
terraform output bucket_arn    # → use in IRSA IAM policy for PutObject + ListBucket
```

## Multi-environment

For staging and production, pass `var.env` at apply time:

```bash
terraform apply -var="env=staging"   # creates csv-processor-uploads-staging
terraform apply -var="env=prod"      # creates csv-processor-uploads-prod
```

Update `ansible/group_vars/all.yaml` and `helm/csv-processor/values-custom.yaml`
with the corresponding bucket name for each environment.
