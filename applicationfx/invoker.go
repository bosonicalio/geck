package applicationfx

import (
	"log/slog"
	"runtime"

	"github.com/hadroncorp/geck/application"
)

func logAppStart(logger *slog.Logger, config application.Config) {
	logger.Info("starting application",
		slog.String("name", config.Name),
		slog.String("environment", config.Environment.String()),
		slog.String("version", config.Version.String()),
		slog.Group("runtime",
			slog.Int("cpus", runtime.NumCPU()),
			slog.Group("go",
				slog.String("version", runtime.Version()),
				slog.String("os", runtime.GOOS),
				slog.String("arch", runtime.GOARCH),
			),
		),
	)
}
