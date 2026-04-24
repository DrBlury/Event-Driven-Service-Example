package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"drblury/event-driven-service/internal/domain"
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
		return nil, errors.New("server configuration is required")
	}
	if routerCfg == nil {
		return nil, errors.New("router configuration is required")
	}

	apiHandler := apihandler.NewAPIHandler(appLogic, info, log, serverCfg.BaseURL, serverCfg.DocsTemplatePath)

	handler := gen.HandlerFromMux(apiHandler, nil)
	handler = otelhttp.NewHandler(handler, "/")

	swagger, err := gen.GetSwagger()
	if err != nil {
		log.Error("failed to get swagger", "error", err)
		return nil, err
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
		OnStart: func(context.Context) error {
			errChan = make(chan error, 1)
			if err := srv.Start(errChan); err != nil {
				return err
			}

			log.With("address", srv.Address()).Info("server started")

			go func() {
				err := <-errChan
				if err == nil || errors.Is(err, http.ErrServerClosed) {
					return
				}
				log.Error("server stopped unexpectedly", "error", err)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := srv.Shutdown(ctx); err != nil {
				log.Error("server shutdown error", "error", err)
				return err
			}
			return nil
		},
	})
}
