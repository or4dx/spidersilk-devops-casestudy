resource "aws_s3_bucket_lifecycle_configuration" "csv_lifecycle" {
  bucket = aws_s3_bucket.csv_uploads.id

  rule {
    id     = "processed-files-archival"
    status = "Enabled"

    # Scoped to the processed/ prefix- the only path the Go app writes to.
    # Other prefixes added in the future (e.g. raw/, audit/) are unaffected.
    filter {
      prefix = "processed/"
    }

    # Day 0–30: STANDARD- hot storage, full retrieval speed.
    # Files are most likely to be re-accessed shortly after upload for review.

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
      # Day 30–90: STANDARD_IA (Infrequent Access).
      # ~58% cheaper than STANDARD for storage; retrieval fee applies.
      # Appropriate once the initial review window has passed.
    }

    transition {
      days          = 90
      storage_class = "GLACIER_IR"
      # Day 90–365: Glacier Instant Retrieval.
      # ~68% cheaper than STANDARD_IA for storage; millisecond retrieval.
      # Chosen over GLACIER (minutes) because occasional ad-hoc lookups
      # should not require a restore job.
    }

    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"
      # Day 365+: Glacier Deep Archive.
      # Lowest cost tier (~$0.00099/GB/month). 12-hour retrieval SLA.
      # Acceptable for files older than a year- compliance reads are
      # planned, not ad-hoc.
    }

    expiration {
      days = 2555 # ~7 years from upload date.
      # Permanent deletion after the compliance retention window.
      # 7 years covers most financial and operational record-keeping
      # requirements (e.g. SOX, GDPR Article 5(1)(e) storage limitation).
    }
  }
}
