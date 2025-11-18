package app

import (
	"drblury/event-driven-service/internal/events"
	"os"
)

// Run orchestrates the application lifecycle from startup to graceful shutdown.
func Run(cfg *Config, shutdownChannel chan os.Signal) error {
	ctx, stop := createAppContext(shutdownChannel)
	defer stop()

	logger := initializeLogger(ctx, cfg)

	if err := initializeTracing(ctx, logger, cfg); err != nil {
		return err
	}

	if err := initializeMetrics(ctx, cfg, logger); err != nil {
		return err
	}

	db, err := connectToDatabase(ctx, cfg, logger)
	if err != nil {
		return err
	}

	appLogic, err := initializeAppLogic(db, logger)
	if err != nil {
		return err
	}

	eventService, err := events.BuildEventService(ctx, cfg.Events, logger, db, appLogic, cfg.Protoflow)
	if err != nil {
		return err
	}
	// add event producer to app logic so we can emit events from use cases
	appLogic.SetEventProducer(eventService)

	httpServer, err := buildHTTPServer(cfg, appLogic, logger)
	if err != nil {
		return err
	}

	srvErr := make(chan error, 1)
	runHTTPServer(httpServer, cfg, logger, srvErr)
	monitorHTTPServerErrors(ctx, srvErr, logger)

	go events.StartEventService(ctx, eventService, logger)
	go events.RunSignupSimulation(ctx, eventService, cfg.Events)

	<-ctx.Done()

	if err := shutdownHTTPServer(httpServer, logger); err != nil {
		return err
	}

	stop()
	return nil
}
