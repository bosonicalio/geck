package validationfx

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"

	"github.com/hadroncorp/geck/validation"
)

// GoPlaygroundModule is the `uber/fx` module of the [validation] package, offering
// implementations with the third-party `go-playground/validator` package.
var GoPlaygroundModule = fx.Module("geck/validation/go-playground",
	fx.Provide(
		env.ParseAs[validation.ValidatorConfig],
		fx.Annotate(
			validation.NewGoPlaygroundValidator,
			fx.As(new(validation.Validator)),
		),
	),
)
