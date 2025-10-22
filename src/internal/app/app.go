package app

import (
	"context"
	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/internal/server"
	generatedAPI "drblury/event-driven-service/internal/server/generated"
	"drblury/event-driven-service/internal/server/handler/apihandler"
	"drblury/event-driven-service/internal/usecase"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/metrics"
	"drblury/event-driven-service/pkg/router"
	"drblury/event-driven-service/pkg/tracing"
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
	logger := logging.SetLogger(ctx, cfg.Logger)

	// ===== Tracing =====
	err := tracing.NewOtelTracer(ctx, logger, cfg.Tracing)
	if err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		return err
	}

	// ===== Metrics =====
	err = metrics.NewOtelMetrics(ctx, cfg.Metrics, logger)
	if err != nil {
		logger.Error("failed to initialize metrics", "error", err)
		return err
	}

	// ===== Database =====
	db, err := database.NewDatabase(cfg.Database, logger, ctx)
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
	eventService := events.NewService(cfg.Events, logger, db, appLogic, ctx)

	logger.With(
		"brokers", eventService.Conf.KafkaBrokers,
		"consume_queue", eventService.Conf.ConsumeQueue,
		"publish_queue", eventService.Conf.PublishQueue).
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
