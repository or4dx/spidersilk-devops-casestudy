package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/or4dx/csv-processor/internal/csvparser"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	s3Client *s3.Client
	tmpl     *template.Template
	s3Bucket string
	logger   *slog.Logger
}

// New constructs a Handler. s3Client may be nil when S3 is not configured.
func New(s3Client *s3.Client, tmpl *template.Template, s3Bucket string, logger *slog.Logger) *Handler {
	return &Handler{
		s3Client: s3Client,
		tmpl:     tmpl,
		s3Bucket: s3Bucket,
		logger:   logger,
	}
}

// PageData is the template context passed to every render call.
type PageData struct {
	Error          string
	S3Warning      string
	Result         *csvparser.CSVResult
	ProcessedFiles []FileInfo
}

// FileInfo represents a single object returned from S3 ListObjectsV2.
type FileInfo struct {
	Key          string
	LastModified time.Time
	Size         int64
}

func (h *Handler) render(w http.ResponseWriter, data PageData) {
	if err := h.tmpl.Execute(w, data); err != nil {
		h.logger.Error("template render failed", "error", err)
	}
}

// renderWithError fetches the files list so the page stays useful even on error.
func (h *Handler) renderWithError(w http.ResponseWriter, r *http.Request, msg string, status int) {
	files, s3Warn := h.listProcessedFiles(r.Context())
	w.WriteHeader(status)
	h.render(w, PageData{
		Error:          msg,
		S3Warning:      s3Warn,
		ProcessedFiles: files,
	})
}
