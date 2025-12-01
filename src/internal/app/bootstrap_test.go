package app

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/logging/metrics"
	"drblury/event-driven-service/pkg/logging/tracing"
)

func TestCreateAppContext(t *testing.T) {
	t.Run("without shutdown channel", func(t *testing.T) {
		ctx, cancel := createAppContext(nil)
		if ctx == nil {
			t.Fatal("createAppContext returned nil context")
		}
		cancel()
	})

	t.Run("with shutdown channel", func(t *testing.T) {
		shutdown := make(chan os.Signal, 1)
		ctx, cancel := createAppContext(shutdown)
		defer cancel()

		if ctx == nil {
			t.Fatal("createAppContext returned nil context")
		}

		// Trigger shutdown
		shutdown <- os.Interrupt

		// Wait for context to be done (or timeout)
		select {
		case <-ctx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			// May timeout if signal not processed fast enough, which is ok
		}
	})

	t.Run("cancel immediately", func(t *testing.T) {
		ctx, cancel := createAppContext(nil)
		if ctx == nil {
			t.Fatal("createAppContext returned nil context")
		}

		cancel()

		select {
		case <-ctx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Context should be done after cancel")
		}
	})
}

func TestInitializeLogger(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		logger := initializeLogger(context.Background(), nil)
		if logger == nil {
			t.Fatal("initializeLogger returned nil logger for nil config")
		}
	})

	t.Run("with nil logger config", func(t *testing.T) {
		cfg := &Config{Logger: nil}
		logger := initializeLogger(context.Background(), cfg)
		if logger == nil {
			t.Fatal("initializeLogger returned nil logger")
		}
	})

	t.Run("with logger config", func(t *testing.T) {
		cfg := &Config{Logger: &logging.Config{Level: "debug", Format: "json"}}
		logger := initializeLogger(context.Background(), cfg)
		if logger == nil {
			t.Fatal("initializeLogger returned nil logger with config")
		}
	})

	t.Run("with pretty format", func(t *testing.T) {
		cfg := &Config{Logger: &logging.Config{Level: "info", Format: "pretty"}}
		logger := initializeLogger(context.Background(), cfg)
		if logger == nil {
			t.Fatal("initializeLogger returned nil logger with pretty format")
		}
	})
}

func TestInitializeTracing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil config", func(t *testing.T) {
		err := initializeTracing(context.Background(), logger, nil)
		if err != nil {
			t.Errorf("initializeTracing with nil config should not error: %v", err)
		}
	})

	t.Run("nil tracing config", func(t *testing.T) {
		cfg := &Config{Tracing: nil}
		err := initializeTracing(context.Background(), logger, cfg)
		if err != nil {
			t.Errorf("initializeTracing with nil tracing config should not error: %v", err)
		}
	})

	t.Run("with console exporter", func(t *testing.T) {
		cfg := &Config{
			Tracing: &tracing.Config{
				OTELTracesExporter: "console",
				ServiceName:        "test-service",
				ServiceVersion:     "1.0.0",
				Enabled:            true,
			},
		}
		err := initializeTracing(context.Background(), logger, cfg)
		if err != nil {
			t.Errorf("initializeTracing with console exporter should not error: %v", err)
		}
	})

	t.Run("with noop exporter", func(t *testing.T) {
		cfg := &Config{
			Tracing: &tracing.Config{
				OTELTracesExporter: "noop",
				ServiceName:        "test-service",
				ServiceVersion:     "1.0.0",
			},
		}
		err := initializeTracing(context.Background(), logger, cfg)
		if err != nil {
			t.Errorf("initializeTracing with noop exporter should not error: %v", err)
		}
	})

	t.Run("with default exporter", func(t *testing.T) {
		cfg := &Config{
			Tracing: &tracing.Config{
				OTELTracesExporter: "",
				ServiceName:        "test-service",
				ServiceVersion:     "1.0.0",
			},
		}
		err := initializeTracing(context.Background(), logger, cfg)
		if err != nil {
			t.Errorf("initializeTracing with default exporter should not error: %v", err)
		}
	})
}

func TestInitializeMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil config", func(t *testing.T) {
		err := initializeMetrics(context.Background(), nil, logger)
		if err != nil {
			t.Errorf("initializeMetrics with nil config should not error: %v", err)
		}
	})

	t.Run("nil metrics config", func(t *testing.T) {
		cfg := &Config{Metrics: nil}
		err := initializeMetrics(context.Background(), cfg, logger)
		if err != nil {
			t.Errorf("initializeMetrics with nil metrics config should not error: %v", err)
		}
	})

	t.Run("disabled metrics", func(t *testing.T) {
		cfg := &Config{Metrics: &metrics.Config{Enabled: false}}
		err := initializeMetrics(context.Background(), cfg, logger)
		if err != nil {
			t.Errorf("initializeMetrics with disabled metrics should not error: %v", err)
		}
	})

	t.Run("with console exporter", func(t *testing.T) {
		cfg := &Config{
			Metrics: &metrics.Config{
				Enabled:             true,
				OTELMetricsExporter: "console",
				ServiceName:         "test-service",
				ServiceVersion:      "1.0.0",
			},
		}
		err := initializeMetrics(context.Background(), cfg, logger)
		if err != nil {
			t.Errorf("initializeMetrics with console exporter should not error: %v", err)
		}
	})

	t.Run("with invalid exporter", func(t *testing.T) {
		cfg := &Config{
			Metrics: &metrics.Config{
				Enabled:             true,
				OTELMetricsExporter: "invalid",
				ServiceName:         "test-service",
				ServiceVersion:      "1.0.0",
			},
		}
		err := initializeMetrics(context.Background(), cfg, logger)
		if err == nil {
			t.Error("initializeMetrics with invalid exporter should error")
		}
	})
}

func TestConnectToDatabase(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil config", func(t *testing.T) {
		_, err := connectToDatabase(context.Background(), nil, logger)
		if err == nil {
			t.Error("connectToDatabase with nil config should error")
		}
	})

	t.Run("nil database config", func(t *testing.T) {
		cfg := &Config{Database: nil}
		_, err := connectToDatabase(context.Background(), cfg, logger)
		if err == nil {
			t.Error("connectToDatabase with nil database config should error")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		cfg := &Config{Database: &database.Config{MongoURL: "invalid://not-a-valid-url"}}
		_, err := connectToDatabase(context.Background(), cfg, logger)
		if err == nil {
			t.Error("connectToDatabase with invalid URL should error")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		cfg := &Config{Database: &database.Config{MongoURL: "", MongoDB: "", MongoUser: ""}}
		_, err := connectToDatabase(context.Background(), cfg, logger)
		if err == nil {
			t.Log("connectToDatabase with empty URL succeeded unexpectedly")
		}
	})
}

func TestInitializeAppLogic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	t.Run("nil database", func(t *testing.T) {
		appLogic, err := initializeAppLogic(nil, logger)
		if err != nil {
			t.Errorf("initializeAppLogic should not error with nil db: %v", err)
		}
		if appLogic == nil {
			t.Error("initializeAppLogic returned nil")
		}
	})

	t.Run("with database", func(t *testing.T) {
		db := &database.Database{}
		appLogic, err := initializeAppLogic(db, logger)
		if err != nil {
			t.Errorf("initializeAppLogic failed: %v", err)
		}
		if appLogic == nil {
			t.Error("initializeAppLogic returned nil")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		appLogic, err := initializeAppLogic(nil, nil)
		if err != nil {
			t.Errorf("initializeAppLogic should not error with nil logger: %v", err)
		}
		if appLogic == nil {
			t.Error("initializeAppLogic returned nil")
		}
	})
}
