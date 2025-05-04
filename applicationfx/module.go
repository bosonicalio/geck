package applicationfx

import (
	"go.uber.org/fx"

	"github.com/tesserical/geck/application"
)

// Module is the `uber/fx` module of the [application] package.
var Module = fx.Module("geck/application",
	fx.Provide(
		application.New,
	),
	fx.Invoke(
		logAppStart,
	),
)
