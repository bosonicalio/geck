package httpfx

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"

	geckhttp "github.com/tesserical/geck/transport/http"
)

// ServerModule is the `uber/fx` module of the [geckhttp] package, aimed for HTTP servers.
//
// This module uses `labstack/echo` as HTTP framework for internal operations.
var ServerModule = fx.Module("geck/transport/http/server",
	fx.Provide(
		env.ParseAs[geckhttp.ServerConfig],
		geckhttp.NewEchoServer,
	),
	fx.Invoke(
		registerServerEndpoints,
		startServer,
	),
)
