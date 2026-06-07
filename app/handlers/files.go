package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Index serves GET / — upload form plus the previously processed files list.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	files, s3Warn := h.listProcessedFiles(r.Context())
	h.render(w, PageData{
		ProcessedFiles: files,
		S3Warning:      s3Warn,
	})
}

// listProcessedFiles calls S3 ListObjectsV2 with the "processed/" prefix.
// On any failure it logs, returns an empty slice, and surfaces a warning string
// so the page continues to render rather than returning an error.
func (h *Handler) listProcessedFiles(ctx context.Context) ([]FileInfo, string) {
	if h.s3Client == nil {
		return nil, "S3 not configured — previously processed files are unavailable."
	}

	out, err := h.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(h.s3Bucket),
		Prefix: aws.String("processed/"),
	})
	if err != nil {
		h.logger.Error("S3 ListObjectsV2 failed", "bucket", h.s3Bucket, "error", err)
		return nil, "Unable to reach S3 — previously processed files temporarily unavailable."
	}

	files := make([]FileInfo, 0, len(out.Contents))
	for _, obj := range out.Contents {
		var lastMod time.Time
		if obj.LastModified != nil {
			lastMod = *obj.LastModified
		}
		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}
		files = append(files, FileInfo{
			Key:          aws.ToString(obj.Key),
			LastModified: lastMod,
			Size:         size,
		})
	}
	return files, ""
}
