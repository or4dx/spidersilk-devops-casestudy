output "bucket_name" {
  description = "S3 bucket name. Use this value in Ansible group_vars (s3_bucket) and Helm values (config.s3Bucket)."
  value       = aws_s3_bucket.csv_uploads.bucket
}

output "bucket_arn" {
  description = "S3 bucket ARN. Reference this in the IRSA IAM policy to scope PutObject and ListBucket permissions to this bucket only."
  value       = aws_s3_bucket.csv_uploads.arn
}
