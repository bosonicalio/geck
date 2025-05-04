package httpfx

import (
	"go.uber.org/fx"

	geckhttp "github.com/tesserical/geck/transport/http"
)

// AsController annotates `t` (preferred a builder routine) as a [geckhttp.Controller] and
// adds it to the HTTP controller registry.
//
// The HTTP registry is used by the `geck` HTTP server module, which eventually will call
// each of the registered [geckhttp.Controller.RegisterEndpoints], exposing all registered endpoints through
// the running HTTP server.
//
// This annotation only works for `uber/fx` providers.
func AsController(t any) any {
	return fx.Annotate(
		t,
		fx.As(new(geckhttp.Controller)),
		fx.ResultTags(`group:"http_controllers"`),
	)
}
