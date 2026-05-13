package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"
	"drblury/event-driven-service/internal/server"
	gen "drblury/event-driven-service/internal/server/gen"
	"drblury/event-driven-service/internal/server/handler/apihandler"
	"drblury/event-driven-service/internal/usecase"

	"github.com/drblury/apiweaver/router"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
)

// buildHTTPServer assembles the HTTP handler stack and returns a configured server.
func buildHTTPServer(
	info *domain.Info,
	routerCfg *router.Config,
	serverCfg *server.Config,
	appLogic *usecase.AppLogic,
	logger *slog.Logger,
) (*server.Server, error) {
	log := fallbackLogger(logger)
	if serverCfg == nil {
		return nil, observability.Builder(context.Background(), "app.http_server", "server_config_missing").
			Public("server configuration is required").
			New("server configuration is required")
	}
	if routerCfg == nil {
		return nil, observability.Builder(context.Background(), "app.http_server", "router_config_missing").
			Public("router configuration is required").
			New("router configuration is required")
	}

	apiHandler := apihandler.NewAPIHandler(appLogic, info, log, serverCfg.BaseURL, serverCfg.DocsTemplatePath)

	handler := gen.HandlerFromMux(apiHandler, nil)
	handler = otelhttp.NewHandler(handler, "/")

	swagger, err := gen.GetSwagger()
	if err != nil {
		wrapped := observability.Builder(context.Background(), "app.http_server", "swagger_load_failed").
			Public("OpenAPI specification could not be loaded").
			Wrap(err)
		log.Error("failed to get swagger", "error", wrapped)
		return nil, wrapped
	}

	options := []router.Option{
		router.WithLogger(log),
		router.WithConfig(*routerCfg),
		router.WithSwagger(swagger),
	}

	r := router.New(handler, options...)
	return server.NewServer(serverCfg, r), nil
}

func registerHTTPServerLifecycle(lc fx.Lifecycle, srv *server.Server, cfg *server.Config, logger *slog.Logger) {
	if srv == nil || cfg == nil {
		return
	}

	log := fallbackLogger(logger)
	var errChan chan error

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errChan = make(chan error, 1)
			if err := srv.Start(errChan); err != nil {
				wrapped := observability.Builder(ctx, "app.http_server", "server_start_failed").
					Public("HTTP server could not start").
					Wrapf(err, "start HTTP server on %s", cfg.Address)
				observability.Logger(ctx, log).Error("failed to start HTTP server", "error", wrapped, "address", cfg.Address)
				return wrapped
			}

			observability.Logger(ctx, log).With("address", srv.Address()).Info("server started")

			go func() {
				err := <-errChan
				if err == nil || errors.Is(err, http.ErrServerClosed) {
					return
				}
				wrapped := observability.Builder(context.Background(), "app.http_server", "server_stopped_unexpectedly").
					Public("HTTP server stopped unexpectedly").
					Wrap(err)
				log.Error("server stopped unexpectedly", "error", wrapped)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := srv.Shutdown(ctx); err != nil {
				wrapped := observability.Builder(ctx, "app.http_server", "server_shutdown_failed").
					Public("HTTP server could not shut down cleanly").
					Wrap(err)
				observability.Logger(ctx, log).Error("server shutdown error", "error", wrapped)
				return wrapped
			}
			return nil
		},
	})
}
