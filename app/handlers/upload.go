package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/or4dx/csv-processor/internal/csvparser"
)

// maxUploadSize caps multipart form memory at 10 MB.
// Files exceeding this are rejected before parsing to avoid memory exhaustion.
const maxUploadSize = 10 << 20

// Upload handles POST /upload: parse multipart form, parse CSV, upload to S3,
// render results. Every error path renders the template with an error banner —
// no panics, no bare http.Error plain-text responses.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.renderWithError(w, r, "File too large — maximum upload size is 10 MB.", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("csv_file")
	if err != nil {
		h.renderWithError(w, r, "No file received — please select a CSV file and try again.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Buffer the upload so we can pass it to both the CSV parser and S3 PutObject.
	// Both consumers need to read from the start; io.Reader is forward-only.
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		h.renderWithError(w, r, "Failed to read the uploaded file.", http.StatusInternalServerError)
		return
	}

	result, err := csvparser.ParseCSV(bytes.NewReader(buf.Bytes()))
	if err != nil {
		h.renderWithError(w, r, fmt.Sprintf("Could not parse CSV: %s.", err), http.StatusBadRequest)
		return
	}

	s3Warning := h.uploadToS3(r.Context(), bytes.NewReader(buf.Bytes()), header.Filename, int64(buf.Len()))

	files, listWarning := h.listProcessedFiles(r.Context())
	if s3Warning == "" {
		s3Warning = listWarning
	}

	h.render(w, PageData{
		Result:         result,
		ProcessedFiles: files,
		S3Warning:      s3Warning,
	})
}

// uploadToS3 stores the original CSV under processed/<RFC3339>_<filename>.
// Returns a non-empty warning string on failure so the caller can surface it
// in the UI without treating it as a fatal error.
func (h *Handler) uploadToS3(ctx context.Context, body io.Reader, filename string, size int64) string {
	if h.s3Client == nil {
		return "S3 not configured — file was not stored."
	}

	key := fmt.Sprintf("processed/%s_%s", time.Now().UTC().Format(time.RFC3339), filename)

	_, err := h.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(h.s3Bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("text/csv"),
	})
	if err != nil {
		h.logger.Error("S3 PutObject failed", "key", key, "error", err)
		return "Could not store file in S3 — your parsed results are still shown below."
	}

	h.logger.Info("uploaded to S3", "key", key, "bytes", size)
	return ""
}
