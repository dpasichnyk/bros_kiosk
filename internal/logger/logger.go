package logger

import (
	"log/slog"
	"os"
)

// Setup configures the default logger to write JSON to stdout.
func Setup() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}
