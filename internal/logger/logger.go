package logger

import (
	"log/slog"
	"os"

	"github.com/jacaudi/tempest_influx/internal/config"
)

// AppLogger wraps slog.Logger to provide structured logging
type AppLogger struct {
	*slog.Logger
}

// New creates a new structured logger based on configuration
func New(cfg *config.Config) *AppLogger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if cfg.Debug {
		opts.Level = slog.LevelDebug
	}

	// Use JSON handler for production, text handler for development
	if cfg.Debug {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return &AppLogger{Logger: logger}
}
