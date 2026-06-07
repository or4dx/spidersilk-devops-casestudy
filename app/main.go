package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/or4dx/csv-processor/handlers"
)

// Config holds all runtime configuration sourced from environment variables.
// Values flow in via Kubernetes ConfigMap/Secret → Helm values → Ansible.
type Config struct {
	Port        string // PORT, default "8080"
	S3Bucket    string // S3_BUCKET — empty means S3 is disabled
	AWSRegion   string // AWS_REGION, default "us-east-1"
	LogLevel    string // LOG_LEVEL: debug | info | warn | error
	TemplateDir string // TEMPLATE_DIR, default "templates"
	StaticDir   string // STATIC_DIR, default "static"
}

func loadConfig() Config {
	return Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		AWSRegion:   getEnvOrDefault("AWS_REGION", "us-east-1"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
		TemplateDir: getEnvOrDefault("TEMPLATE_DIR", "templates"),
		StaticDir:   getEnvOrDefault("STATIC_DIR", "static"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	cfg := loadConfig()

	logger := buildLogger(cfg.LogLevel)

	// Fail fast: if the template can't load at startup, the container should
	// crash immediately rather than serve 500s on every request.
	tmplPath := filepath.Join(cfg.TemplateDir, "index.html")
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"humanSize": func(b int64) string {
			const kb, mb = int64(1024), int64(1024 * 1024)
			switch {
			case b >= mb:
				return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
			case b >= kb:
				return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
			default:
				return fmt.Sprintf("%d B", b)
			}
		},
	}).ParseFiles(tmplPath)
	if err != nil {
		logger.Error("failed to load template", "path", tmplPath, "error", err)
		os.Exit(1)
	}

	// Initialise S3 client only when S3_BUCKET is set.
	// Absent bucket: uploads are skipped and the file list renders empty.
	// Credentials are never hardcoded — config.LoadDefaultConfig respects
	// env vars (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY), IRSA, and EC2
	// instance profiles in priority order.
	var s3Client *s3.Client
	if cfg.S3Bucket == "" {
		logger.Warn("S3_BUCKET not set — uploads and file listing disabled")
	} else {
		awsCfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.AWSRegion),
		)
		if err != nil {
			logger.Error("failed to load AWS config", "error", err)
			os.Exit(1)
		}
		s3Client = s3.NewFromConfig(awsCfg)
		logger.Info("S3 client initialised", "bucket", cfg.S3Bucket, "region", cfg.AWSRegion)
	}

	h := handlers.New(s3Client, tmpl, cfg.S3Bucket, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.Index)
	mux.HandleFunc("POST /upload", h.Upload)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Serve static assets from disk. In Kubernetes, Nginx intercepts /static/
	// before traffic reaches Go, so this handler is only active locally.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: block until SIGINT or SIGTERM, then give in-flight
	// requests 10 seconds to complete before the process exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}

func buildLogger(level string) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
