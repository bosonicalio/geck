package logging

import (
	"log/slog"
	"os"
)

// NewSlogLogger allocates a new [slog.Logger] instance with default configurations.
func NewSlogLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
}
