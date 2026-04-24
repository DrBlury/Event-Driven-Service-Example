package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/logging/metrics"
	"drblury/event-driven-service/pkg/logging/tracing"

	"github.com/drblury/protoflow"
	"go.uber.org/fx"
	"google.golang.org/protobuf/proto"
)

func TestSplitConfig(t *testing.T) {
	cfg := minimalConfig()
	cfg.Logger = &logging.Config{Level: "debug", Format: "json"}
	cfg.Tracing = &tracing.Config{Enabled: true}
	cfg.Metrics = &metrics.Config{Enabled: true}

	sections := splitConfig(cfg)
	if sections.Info != cfg.Info || sections.Router != cfg.Router || sections.Server != cfg.Server {
		t.Fatal("expected splitConfig to preserve top-level references")
	}
	if sections.Database != cfg.Database || sections.Logger != cfg.Logger || sections.Tracing != cfg.Tracing || sections.Metrics != cfg.Metrics {
		t.Fatal("expected splitConfig to preserve nested references")
	}
	if sections.Protoflow != cfg.Protoflow || sections.Events != cfg.Events {
		t.Fatal("expected splitConfig to preserve event references")
	}
}

func TestProvideLogger(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		logger := provideLogger(nil)
		if logger == nil {
			t.Fatal("provideLogger returned nil")
		}
	})

	t.Run("configured logger", func(t *testing.T) {
		logger := provideLogger(&logging.Config{Level: "debug", Format: "json"})
		if logger == nil {
			t.Fatal("provideLogger returned nil")
		}
	})
}

func TestProvideFXLogger(t *testing.T) {
	fxLogger := provideFXLogger(nil)
	if fxLogger == nil {
		t.Fatal("expected fx logger")
	}

	fxLogger = provideFXLogger(&logging.Config{Level: "debug", Format: "json"})
	if fxLogger == nil {
		t.Fatal("expected fx logger")
	}
}

func TestProvideAppLogicConfiguresEventPublishing(t *testing.T) {
	producer := &bootstrapMockProducer{}
	cfg := &events.Config{ExampleConsumeQueue: "example-topic"}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	logic, err := provideAppLogic(nil, logger, cfg, producer)
	if err != nil {
		t.Fatalf("provideAppLogic returned error: %v", err)
	}
	if logic == nil {
		t.Fatal("expected app logic")
	}
	if logic.ExampleTopic() != cfg.ExampleConsumeQueue {
		t.Fatalf("topic = %q, want %q", logic.ExampleTopic(), cfg.ExampleConsumeQueue)
	}
}

func TestRegisterShutdownChannelHookTriggersShutdown(t *testing.T) {
	shutdownChan := make(chan os.Signal, 1)
	app := fx.New(
		fx.Supply(ShutdownChannel(shutdownChan)),
		fx.Supply(slog.New(slog.NewTextHandler(os.Stderr, nil))),
		fx.Invoke(registerShutdownChannelHook),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx.New returned error: %v", err)
	}

	startCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("app.Start returned error: %v", err)
	}

	shutdownChan <- os.Interrupt

	select {
	case <-app.Wait():
	case <-time.After(time.Second):
		t.Fatal("expected shutdown signal from fx app")
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		t.Fatalf("app.Stop returned error: %v", err)
	}
}

func TestRegisterTelemetryHooksRejectsInvalidMetricsExporter(t *testing.T) {
	app := fx.New(
		fx.Supply((*tracing.Config)(nil)),
		fx.Supply(&metrics.Config{
			Enabled:             true,
			OTELMetricsExporter: "invalid",
			ServiceName:         "test-service",
			ServiceVersion:      "1.0.0",
		}),
		fx.Supply(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))),
		fx.Invoke(registerTelemetryHooks),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx.New returned error: %v", err)
	}

	if err := app.Start(context.Background()); err == nil {
		t.Fatal("expected startup error")
	}
}

func TestProvideDatabaseUsesLifecycleStartContext(t *testing.T) {
	app := fx.New(
		fx.Supply(&database.Config{
			MongoURL: "mongodb://localhost:27017/?serverSelectionTimeoutMS=5000",
			MongoDB:  "testdb",
		}),
		fx.Supply(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))),
		fx.Provide(provideDatabase),
		fx.Invoke(func(*database.Database) {}),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx.New returned error: %v", err)
	}

	startCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := app.Start(startCtx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("app.Start error = %v, want context canceled", err)
	}
}

type bootstrapMockProducer struct{}

func (bootstrapMockProducer) PublishProto(context.Context, string, proto.Message, protoflow.Metadata) error {
	return nil
}
