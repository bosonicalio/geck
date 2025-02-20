package identifierfx

import (
	"go.uber.org/fx"

	"github.com/hadroncorp/geck/persistence/identifier"
)

// KSUIDModule is the `uber/fx` module of the [identifier] package, offering implementations
// using the `segmentio/ksuid` package.
var KSUIDModule = fx.Module("geck/persistence/identifier",
	fx.Provide(
		fx.Annotate(
			func() identifier.FactoryKSUID {
				return identifier.FactoryKSUID{}
			},
			fx.As(new(identifier.Factory)),
		),
	),
)
