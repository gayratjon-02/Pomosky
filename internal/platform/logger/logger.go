// Package logger builds the application's structured logger.
package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON logger in production and a readable one otherwise.
func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	opts.Level = slog.LevelDebug
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
