package loggingfx

import (
	"go.uber.org/fx"

	"github.com/hadroncorp/geck/observability/logging"
)

// SlogModule is the `uber/fx` module of the [logging] package, using
// stdlib `slog` package for concrete implementations.
var SlogModule = fx.Module("geck/observability/logging/slog",
	fx.Provide(
		logging.NewSlogLogger,
	),
)
