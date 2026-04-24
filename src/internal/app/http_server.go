package app

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"drblury/event-driven-service/internal/server"
	gen "drblury/event-driven-service/internal/server/gen"
	"drblury/event-driven-service/internal/server/handler/apihandler"
	"drblury/event-driven-service/internal/usecase"

	"github.com/drblury/apiweaver/router"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
)

// buildHTTPServer assembles the HTTP handler stack and returns a configured server.
func buildHTTPServer(cfg *Config, appLogic *usecase.AppLogic, logger *slog.Logger) (*server.Server, error) {
	if cfg == nil {
		return nil, errors.New("application configuration is required")
	}
	if cfg.Server == nil {
		return nil, errors.New("server configuration is required")
	}
	if cfg.Router == nil {
		return nil, errors.New("router configuration is required")
	}

	apiHandler := apihandler.NewAPIHandler(appLogic, cfg.Info, logger, cfg.Server.BaseURL, cfg.Server.DocsTemplatePath)

	handler := gen.HandlerFromMux(apiHandler, nil)
	handler = otelhttp.NewHandler(handler, "/")

	swagger, err := gen.GetSwagger()
	if err != nil {
		logger.Error("failed to get swagger", "error", err)
		return nil, err
	}

	options := []router.Option{
		router.WithLogger(logger),
		router.WithConfig(*cfg.Router),
		router.WithSwagger(swagger),
	}

	r := router.New(handler, options...)
	return server.NewServer(cfg.Server, r), nil
}

func registerHTTPServerLifecycle(lc fx.Lifecycle, srv *server.Server, cfg *server.Config, logger *slog.Logger) {
	if srv == nil || cfg == nil {
		return
	}

	var errChan chan error

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			errChan = make(chan error, 1)
			if err := runHTTPServer(srv, &Config{Server: cfg}, logger, errChan); err != nil {
				return err
			}
			monitorHTTPServerErrors(context.Background(), errChan, logger)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return shutdownHTTPServerWithContext(ctx, srv, logger)
		},
	})
}

// runHTTPServer starts the HTTP server asynchronously and forwards fatal errors to the provided channel.
func runHTTPServer(srv *server.Server, cfg *Config, logger *slog.Logger, errChan chan<- error) error {
	if srv == nil || cfg == nil || cfg.Server == nil {
		return nil
	}

	if err := srv.Start(errChan); err != nil {
		return err
	}

	if logger != nil {
		logger.With("address", cfg.Server.Address).Info("server started")
	}
	return nil
}

// monitorHTTPServerErrors reports unexpected server termination.
func monitorHTTPServerErrors(ctx context.Context, errChan <-chan error, logger *slog.Logger) {
	if errChan == nil {
		return
	}

	go func() {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			if err == nil || errors.Is(err, http.ErrServerClosed) {
				return
			}
			if logger != nil {
				logger.Error("server stopped unexpectedly", "error", err)
			}
		}
	}()
}

// shutdownHTTPServer gracefully terminates HTTP handling.
func shutdownHTTPServer(srv *server.Server, logger *slog.Logger) error {
	return shutdownHTTPServerWithContext(context.Background(), srv, logger)
}

func shutdownHTTPServerWithContext(ctx context.Context, srv *server.Server, logger *slog.Logger) error {
	if srv == nil {
		return nil
	}

	if err := srv.Shutdown(ctx); err != nil {
		if logger != nil {
			logger.Error("server shutdown error", "error", err)
		}
		return err
	}
	return nil
}
