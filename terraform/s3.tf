resource "aws_s3_bucket" "csv_uploads" {
  bucket = "${var.bucket_name_prefix}-${var.env}"

  tags = {
    Name        = "${var.bucket_name_prefix}-${var.env}"
    Environment = var.env
    ManagedBy   = "terraform"
  }
}

resource "aws_s3_bucket_versioning" "csv_uploads" {
  bucket = aws_s3_bucket.csv_uploads.id

  versioning_configuration {
    status = "Enabled"
    # Versioning protects against accidental overwrites. The app writes
    # RFC3339-timestamped keys (processed/2024-01-15T10:30:00Z_file.csv)
    # so collisions are unlikely, but versioning adds a safety net at no
    # extra storage cost for small CSV files.
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "csv_uploads" {
  bucket = aws_s3_bucket.csv_uploads.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
      # AES256 (SSE-S3) chosen over aws:kms:
      #   - Free - no per-request charges, no CMK cost ($1/key/month for KMS)
      #   - Sufficient for CSV inventory data with no PII/HIPAA/PCI classification
      #   - KMS is justified when per-object CloudTrail audit logs or custom key
      #     rotation policies are a compliance requirement
    }
  }
}

resource "aws_s3_bucket_public_access_block" "csv_uploads" {
  bucket = aws_s3_bucket.csv_uploads.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
  # All four flags enabled- belt-and-suspenders. The Go app uses server-side
  # PutObject calls authenticated via IRSA. The bucket must never be reachable
  # from the public internet regardless of any future policy changes.
}

# CORS intentionally omitted.
# The Go app handles all S3 interaction server-side (PutObject via the AWS SDK
# with IRSA credentials). The browser never communicates with S3 directly,
# so CORS headers serve no purpose in this architecture.
#
# If the design moves to presigned-URL direct browser uploads, add a CORS
# configuration here with AllowedOrigins scoped to the app's domain.
