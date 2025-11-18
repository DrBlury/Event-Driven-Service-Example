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
)

// buildHTTPServer assembles the HTTP handler stack and returns a configured server.
func buildHTTPServer(cfg *Config, appLogic *usecase.AppLogic, logger *slog.Logger) (*server.Server, error) {
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

// runHTTPServer starts the HTTP server asynchronously and forwards fatal errors to the provided channel.
func runHTTPServer(srv *server.Server, cfg *Config, logger *slog.Logger, errChan chan<- error) {
	if srv == nil || cfg == nil || cfg.Server == nil {
		return
	}

	go func() {
		logger.With("address", cfg.Server.Address).Info("server started!")
		if err := srv.ListenAndServe(); errChan != nil {
			errChan <- err
		}
	}()
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
			logger.Error("server stopped unexpectedly", "error", err)
		}
	}()
}

// shutdownHTTPServer gracefully terminates HTTP handling.
func shutdownHTTPServer(srv *server.Server, logger *slog.Logger) error {
	if srv == nil {
		return nil
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("server shutdown error", "error", err)
		return err
	}
	return nil
}
