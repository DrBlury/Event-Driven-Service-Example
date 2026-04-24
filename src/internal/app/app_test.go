package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/pkg/logging"

	"github.com/drblury/apiweaver/router"
	"github.com/drblury/protoflow"
)

func TestLoadConfigFromMetadata(t *testing.T) {
	metadata := Metadata{
		Version:     "1.2.3",
		BuildDate:   "2026-04-24",
		Description: "fx-test",
		CommitHash:  "abc123",
		CommitDate:  "2026-04-24T00:00:00Z",
	}

	cfg, err := loadConfigFromMetadata(metadata)
	if err != nil {
		t.Fatalf("loadConfigFromMetadata returned error: %v", err)
	}
	if cfg == nil || cfg.Info == nil {
		t.Fatal("expected populated config and info")
	}
	if cfg.Info.Version != metadata.Version || cfg.Info.BuildDate != metadata.BuildDate {
		t.Fatalf("metadata not propagated to config: %+v", cfg.Info)
	}
	if cfg.Info.Details != metadata.Description || cfg.Info.CommitHash != metadata.CommitHash || cfg.Info.CommitDate != metadata.CommitDate {
		t.Fatalf("commit metadata not propagated to config: %+v", cfg.Info)
	}
}

func TestNewFromConfigStartFailsWithoutDatabaseConfig(t *testing.T) {
	cfg := minimalConfig()
	cfg.Database = nil

	app := NewFromConfig(cfg, nil)
	err := app.Start(context.Background())
	if err == nil {
		t.Fatal("expected startup error")
	}
	if !strings.Contains(err.Error(), "database configuration is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewFromConfigStartFailsWithInvalidDatabaseURL(t *testing.T) {
	cfg := minimalConfig()
	cfg.Database = &database.Config{
		MongoURL:      "invalid://not-a-valid-url",
		MongoDB:       "testdb",
		MongoUser:     "user",
		MongoPassword: "pass",
	}

	app := NewFromConfig(cfg, nil)
	if err := app.Start(context.Background()); err == nil {
		t.Fatal("expected startup error")
	}
}

func minimalConfig() *Config {
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
			Address: "127.0.0.1:0",
		},
		Router: &router.Config{Timeout: 30 * time.Second},
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
