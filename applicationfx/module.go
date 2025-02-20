package applicationfx

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"

	"github.com/hadroncorp/geck/application"
)

// Module is the `uber/fx` module of the [application] package.
var Module = fx.Module("geck/application",
	fx.Provide(
		env.ParseAs[application.Config],
	),
	fx.Invoke(
		logAppStart,
	),
)
