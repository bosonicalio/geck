package eventfx

import (
	"go.uber.org/fx"

	"github.com/tesserical/geck/event"
	"github.com/tesserical/geck/persistence/identifier"
)

var PublisherModule = fx.Module("geck/event",
	fx.Provide(
		fx.Annotate(
			identifier.NewUUIDFactory,
			fx.As(new(identifier.Factory)),
			fx.ResultTags(`name:"message_id_factory"`),
		),
		fx.Annotate(
			event.NewStreamPublisher,
			fx.As(new(event.Publisher)),
			fx.ParamTags("", `name:"message_id_factory"`),
		),
	),
)
