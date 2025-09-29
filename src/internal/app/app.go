package app

import (
	"context"
	"drblury/poc-event-signup/internal/database"
	"drblury/poc-event-signup/internal/events"
	"drblury/poc-event-signup/internal/server"
	generatedAPI "drblury/poc-event-signup/internal/server/generated"
	"drblury/poc-event-signup/internal/server/handler/apihandler"
	"drblury/poc-event-signup/internal/usecase"
	"drblury/poc-event-signup/pkg/logging"
	"drblury/poc-event-signup/pkg/router"
	"os"
	"os/signal"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Run runs the app
// nolint: funlen
func Run(cfg *Config, shutdownChannel chan os.Signal) error {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// ===== Logger =====
	logger := logging.SetLogger(cfg.Logger)

	// ===== Database =====
	db, err := database.NewDatabase(cfg.Database, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return err
	}

	// ===== App Logic =====
	appLogic := usecase.NewAppLogic(
		db,
		logger,
	)

	// ===== HTTP Handlers =====
	apiHandler := apihandler.NewAPIHandler(appLogic, cfg.Info, logger, cfg.Server.BaseURL)

	// ===== HTTP Router =====
	handler := generatedAPI.HandlerFromMux(apiHandler, nil)

	// wrap with otel middleware
	handler = otelhttp.NewHandler(handler, "/")

	swagger, err := generatedAPI.GetSwagger()
	if err != nil {
		logger.Error("failed to get swagger", "error", err)
		return err
	}
	r := router.New(handler, cfg.Router, logger, swagger)

	// ===== Server ===s==
	srv := server.NewServer(cfg.Server, r)

	srvErr := make(chan error, 1)
	go func() {
		logger.With("address", cfg.Server.Address).Info("server started!")
		srvErr <- srv.ListenAndServe()
	}()

	// ===== Event Handling =====
	eventService := events.NewService(cfg.Events, logger, ctx)

	logger.With(
		"brokers", eventService.Conf.KafkaBrokers,
		"consume_topic", eventService.Conf.ConsumeTopic,
		"publish_topic", eventService.Conf.PublishTopic).
		Info("starting event service")

	// Wait for interruption.
	<-ctx.Done()

	// Stop receiving signal notifications as soon as possible.
	err = srv.Shutdown(context.Background())
	if err != nil {
		logger.Error("server shutdown error", "error", err)
		return err
	}
	stop()

	return nil
}
