package httpfx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/hadroncorp/geck/application"
	geckhttp "github.com/hadroncorp/geck/transport/http"
)

type startServerDeps struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *slog.Logger
	Echo      *echo.Echo
	Config    geckhttp.ServerConfig
}

func startServer(deps startServerDeps) {
	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				deps.Logger.InfoContext(ctx, "starting http server",
					slog.String("addr", deps.Config.Address),
				)
				err := geckhttp.StartServer(deps.Echo, deps.Config)
				if errors.Is(err, http.ErrServerClosed) {
					return
				} else if err != nil {
					deps.Logger.ErrorContext(ctx, "failed during http server execution",
						slog.String("error", err.Error()),
					)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			deps.Logger.InfoContext(ctx, "stopping http server")
			return deps.Echo.Shutdown(ctx)
		},
	})
}

type registerServerEndpointsDeps struct {
	fx.In
	Echo        *echo.Echo
	Logger      *slog.Logger
	AppConfig   application.Config
	Controllers []geckhttp.Controller `group:"http_controllers"`
}

func registerServerEndpoints(deps registerServerEndpointsDeps) {
	pathPrefix := geckhttp.RegisterServerEndpoints(deps.Echo, deps.AppConfig, deps.Controllers)
	deps.Logger.Debug("registered http server endpoints",
		slog.String("path_prefix", pathPrefix),
		slog.Int("total_controllers", len(deps.Controllers)),
	)
}
