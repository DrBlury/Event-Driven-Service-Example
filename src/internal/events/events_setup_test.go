package events

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

func TestBuildEventService_NilConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	_, err := BuildEventService(context.Background(), nil, logger, nil, nil, nil)
	if err == nil {
		t.Error("expected error when config is nil")
	}
}

func TestBuildEventService_NilProtoflowConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{}

	_, err := BuildEventService(context.Background(), cfg, logger, nil, nil, nil)
	if err == nil {
		t.Error("expected error when protoflow config is nil")
	}
}

func TestComposeEventMiddlewares(t *testing.T) {
	cfg := &protoflow.Config{
		RetryMaxRetries:      3,
		RetryInitialInterval: 100,
		RetryMaxInterval:     1000,
	}

	middlewares := composeEventMiddlewares(cfg)

	// Expected number of middlewares
	expectedCount := 8
	if len(middlewares) != expectedCount {
		t.Errorf("expected %d middlewares, got %d", expectedCount, len(middlewares))
	}
}

func TestPoisonQueueFilter_UnprocessableError(t *testing.T) {
	filter := poisonQueueFilter()

	// Test with unprocessable error - use ErrUnprocessable sentinel error
	unprocessableErr := protoflow.ErrUnprocessable
	if !filter(unprocessableErr) {
		t.Error("expected filter to return true for UnprocessableEventError")
	}
}

func TestPoisonQueueFilter_ValidationError(t *testing.T) {
	filter := poisonQueueFilter()

	// Test with validation error
	validationErr := domain.ErrValidations{
		Errors: []string{"test: invalid"},
	}
	if !filter(validationErr) {
		t.Error("expected filter to return true for ErrValidations")
	}
}

func TestPoisonQueueFilter_RegularError(t *testing.T) {
	filter := poisonQueueFilter()

	// Test with regular error
	regularErr := errors.New("some error")
	if filter(regularErr) {
		t.Error("expected filter to return false for regular error")
	}
}

func TestStartEventService_NilService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Should not panic when service is nil
	StartEventService(context.Background(), nil, logger)
}

func TestLogEventServiceStartup_NilService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Should not panic when service is nil
	logEventServiceStartup(logger, nil)
}

func TestLogEventServiceStartup_NilConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	// Cannot create Service with nil config directly as fields are not exported
	// Just test with nil service which covers the nil check path
	logEventServiceStartup(logger, nil)
}
