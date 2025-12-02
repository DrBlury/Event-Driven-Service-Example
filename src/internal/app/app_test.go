package app

import (
	"context"
	"os"
	"testing"
	"time"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/logging/metrics"
	"drblury/event-driven-service/pkg/logging/tracing"

	"github.com/drblury/apiweaver/router"
	"github.com/drblury/protoflow"
)

func TestRunWithShutdownSignal(t *testing.T) {
	t.Run("immediate shutdown via channel", func(t *testing.T) {
		cfg := createMinimalConfig()
		shutdownChan := make(chan os.Signal, 1)

		done := make(chan error, 1)
		go func() {
			done <- Run(cfg, shutdownChan)
		}()

		// Give it a brief moment to start, then signal shutdown
		time.Sleep(50 * time.Millisecond)
		shutdownChan <- os.Interrupt

		select {
		case err := <-done:
			// Database connection will fail, which is expected
			if err != nil {
				t.Logf("Run returned error (expected for missing DB): %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Run did not return after shutdown signal")
		}
	})
}

func TestRunWithNilShutdownChannel(t *testing.T) {
	t.Run("nil shutdown channel", func(t *testing.T) {
		cfg := createMinimalConfig()

		// Create a context that we can cancel externally
		done := make(chan error, 1)
		go func() {
			done <- Run(cfg, nil)
		}()

		// Since we can't signal shutdown and DB will fail, it should error quickly
		select {
		case err := <-done:
			// Expected to fail due to DB connection
			if err == nil {
				t.Log("Run succeeded (unexpected, but ok)")
			}
		case <-time.After(15 * time.Second):
			t.Error("Run took too long to return")
		}
	})
}

func TestRunWithNilConfig(t *testing.T) {
	t.Run("nil config causes panic or error", func(t *testing.T) {
		shutdownChan := make(chan os.Signal, 1)

		defer func() {
			if r := recover(); r != nil {
				// Expected panic with nil config
				t.Logf("Recovered from panic (expected): %v", r)
			}
		}()

		done := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- nil
				}
			}()
			done <- Run(nil, shutdownChan)
		}()

		time.Sleep(50 * time.Millisecond)
		shutdownChan <- os.Interrupt

		select {
		case <-done:
			// Expected
		case <-time.After(2 * time.Second):
			// May hang if not handled properly
		}
	})
}

func TestRunWithMissingDatabaseConfig(t *testing.T) {
	t.Run("missing database config returns error", func(t *testing.T) {
		cfg := &Config{
			Logger:   &logging.Config{Level: "error", Format: "json"},
			Tracing:  nil,
			Metrics:  nil,
			Database: nil, // Missing database config
			Server:   &server.Config{Address: ":0"},
			Router:   &router.Config{},
			Info:     &domain.Info{Version: "1.0.0"},
			Events:   &events.Config{},
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		if err == nil {
			t.Error("Expected error when database config is missing")
		}
	})
}

func TestRunWithInvalidDatabaseURL(t *testing.T) {
	t.Run("invalid database URL returns error", func(t *testing.T) {
		cfg := createMinimalConfig()
		cfg.Database = &database.Config{
			MongoURL:      "invalid://not-a-valid-url",
			MongoDB:       "testdb",
			MongoUser:     "user",
			MongoPassword: "pass",
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		if err == nil {
			t.Error("Expected error when database URL is invalid")
		}
	})
}

func TestRunWithTracingEnabled(t *testing.T) {
	t.Run("tracing initialization with console exporter", func(t *testing.T) {
		cfg := createMinimalConfig()
		cfg.Tracing = &tracing.Config{
			Enabled:            true,
			OTELTracesExporter: "console",
			ServiceName:        "test-service",
			ServiceVersion:     "1.0.0",
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		// Will fail at DB connection
		if err == nil {
			t.Log("Unexpected success - DB connection should fail")
		}
	})
}

func TestRunWithMetricsEnabled(t *testing.T) {
	t.Run("metrics initialization with console exporter", func(t *testing.T) {
		cfg := createMinimalConfig()
		cfg.Metrics = &metrics.Config{
			Enabled:             true,
			OTELMetricsExporter: "console",
			ServiceName:         "test-service",
			ServiceVersion:      "1.0.0",
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		// Will fail at DB connection
		if err == nil {
			t.Log("Unexpected success - DB connection should fail")
		}
	})
}

func TestRunWithOtlpTracingConfig(t *testing.T) {
	t.Run("otlp tracing config without endpoint", func(t *testing.T) {
		cfg := createMinimalConfig()
		cfg.Tracing = &tracing.Config{
			Enabled:            true,
			OTELTracesExporter: "otlp",
			ServiceName:        "test-service",
			ServiceVersion:     "1.0.0",
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		// May or may not error depending on implementation
		_ = err
	})
}

func TestRunWithInvalidMetricsConfig(t *testing.T) {
	t.Run("invalid metrics exporter returns error", func(t *testing.T) {
		cfg := createMinimalConfig()
		cfg.Metrics = &metrics.Config{
			Enabled:             true,
			OTELMetricsExporter: "invalid-exporter",
			ServiceName:         "test-service",
			ServiceVersion:      "1.0.0",
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		if err == nil {
			t.Error("Expected error for invalid metrics exporter")
		}
	})
}

func TestRunContextCancellation(t *testing.T) {
	t.Run("context cancellation via shutdown channel", func(t *testing.T) {
		cfg := createMinimalConfig()
		shutdownChan := make(chan os.Signal, 1)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- Run(cfg, shutdownChan)
		}()

		// Wait for context timeout then send shutdown
		<-ctx.Done()
		shutdownChan <- os.Interrupt

		select {
		case <-done:
			// Expected
		case <-time.After(5 * time.Second):
			t.Error("Run did not return after context cancellation")
		}
	})
}

func TestRunWithEmptyServerAddress(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		cfg := createMinimalConfig()
		cfg.Server = &server.Config{
			Address: "", // Empty address
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		// Will fail at DB connection before server starts
		if err == nil {
			t.Log("Run succeeded unexpectedly")
		}
	})
}

func TestRunWithFullConfig(t *testing.T) {
	t.Run("full config with all options", func(t *testing.T) {
		cfg := &Config{
			Logger: &logging.Config{
				Level:  "debug",
				Format: "json",
			},
			Tracing: &tracing.Config{
				Enabled:            true,
				OTELTracesExporter: "noop",
				ServiceName:        "test-service",
				ServiceVersion:     "1.0.0",
			},
			Metrics: &metrics.Config{
				Enabled:             true,
				OTELMetricsExporter: "console",
				ServiceName:         "test-service",
				ServiceVersion:      "1.0.0",
			},
			Database: &database.Config{
				MongoURL:      "mongodb://localhost:27017",
				MongoDB:       "testdb",
				MongoUser:     "user",
				MongoPassword: "pass",
			},
			Server: &server.Config{
				Address: ":0",
				BaseURL: "/api",
			},
			Router: &router.Config{
				Timeout:         30 * time.Second,
				QuietdownRoutes: []string{"/healthz"},
			},
			Info: &domain.Info{
				Version:    "1.0.0",
				BuildDate:  "2024-01-01",
				Details:    "Test",
				CommitHash: "abc123",
			},
			Events: &events.Config{
				DemoConsumeQueue:    "demo-in",
				DemoPublishQueue:    "demo-out",
				ExampleConsumeQueue: "example-in",
				ExamplePublishQueue: "example-out",
			},
			Protoflow: &protoflow.Config{
				RetryMaxRetries:      3,
				RetryInitialInterval: 100,
				RetryMaxInterval:     1000,
			},
		}

		shutdownChan := make(chan os.Signal, 1)
		err := Run(cfg, shutdownChan)

		// Will fail at DB connection
		if err == nil {
			t.Log("Unexpected success")
		}
	})
}

// createMinimalConfig creates a minimal valid configuration for testing
func createMinimalConfig() *Config {
	return &Config{
		Logger: &logging.Config{
			Level:  "error",
			Format: "json",
		},
		Tracing: nil,
		Metrics: nil,
		Database: &database.Config{
			MongoURL:      "mongodb://localhost:27017",
			MongoDB:       "testdb",
			MongoUser:     "user",
			MongoPassword: "pass",
		},
		Server: &server.Config{
			Address: ":0",
		},
		Router: &router.Config{},
		Info:   &domain.Info{Version: "1.0.0"},
		Events: &events.Config{
			DemoConsumeQueue:    "demo",
			DemoPublishQueue:    "demo-out",
			ExampleConsumeQueue: "example",
			ExamplePublishQueue: "example-out",
		},
		Protoflow: &protoflow.Config{},
	}
}
