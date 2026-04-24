package app

import (
	"context"
	"os"

	"drblury/event-driven-service/internal/events"

	"go.uber.org/fx"
)

type Metadata struct {
	Version     string
	BuildDate   string
	Description string
	CommitHash  string
	CommitDate  string
}

// Run starts the Fx application, waits for a shutdown signal, and stops it gracefully.
func Run(cfg *Config, shutdownChannel chan os.Signal) error {
	application := NewFromConfig(cfg, shutdownChannel)

	startCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancel()

	if err := application.Start(startCtx); err != nil {
		return err
	}

	<-application.Wait()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer stopCancel()

	if err := application.Stop(stopCtx); err != nil {
		return err
	}

	if err := application.Err(); err != nil {
		return err
	}

	return nil
}

// New builds the Fx application from compile-time metadata.
func New(metadata Metadata, shutdownChannel chan os.Signal, opts ...fx.Option) *fx.App {
	return newFXApp(
		shutdownChannel,
		append(
			[]fx.Option{
				fx.Supply(metadata),
				fx.Provide(loadConfigFromMetadata),
			},
			opts...,
		)...,
	)
}

// NewFromConfig builds the Fx application from an already loaded config.
func NewFromConfig(cfg *Config, shutdownChannel chan os.Signal, opts ...fx.Option) *fx.App {
	return newFXApp(
		shutdownChannel,
		append(
			[]fx.Option{
				fx.Supply(cfg),
			},
			opts...,
		)...,
	)
}

func newFXApp(shutdownChannel chan os.Signal, extraOpts ...fx.Option) *fx.App {
	appOpts := []fx.Option{
		fx.Provide(
			splitConfig,
			provideLogger,
			provideDatabase,
			events.BuildEventService,
			provideEventProducer,
			provideAppLogic,
			buildHTTPServer,
		),
		fx.Invoke(
			registerTelemetryHooks,
			registerHTTPServerLifecycle,
			events.RegisterLifecycle,
		),
		fx.WithLogger(provideFXLogger),
	}

	if shutdownChannel != nil {
		appOpts = append(
			appOpts,
			fx.Supply(ShutdownChannel(shutdownChannel)),
			fx.Invoke(registerShutdownChannelHook),
		)
	}

	appOpts = append(appOpts, extraOpts...)
	return fx.New(appOpts...)
}

func loadConfigFromMetadata(metadata Metadata) (*Config, error) {
	return LoadConfig(
		metadata.Version,
		metadata.BuildDate,
		metadata.Description,
		metadata.CommitHash,
		metadata.CommitDate,
	)
}
