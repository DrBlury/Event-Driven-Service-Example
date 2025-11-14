package app

import "os"

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

	appLogic := initializeAppLogic(db, logger)

	httpServer, err := buildHTTPServer(cfg, appLogic, logger)
	if err != nil {
		return err
	}

	srvErr := make(chan error, 1)
	runHTTPServer(httpServer, cfg, logger, srvErr)
	monitorHTTPServerErrors(ctx, srvErr, logger)

	eventService, err := buildEventService(ctx, cfg, logger, db, appLogic)
	if err != nil {
		return err
	}

	go startEventService(ctx, eventService, logger)
	go runSignupSimulation(ctx, eventService)

	logEventServiceStartup(logger, eventService)

	<-ctx.Done()

	if err := shutdownHTTPServer(httpServer, logger); err != nil {
		return err
	}

	stop()
	return nil
}
