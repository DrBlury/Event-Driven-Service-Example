package events

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

func TestBuildEventService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil config returns error", func(t *testing.T) {
		_, err := BuildEventService(context.Background(), nil, logger, nil, nil, nil)
		if err == nil {
			t.Error("expected error when config is nil")
		}
		if err.Error() != "events configuration is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("nil protoflow config returns error", func(t *testing.T) {
		cfg := &Config{}
		_, err := BuildEventService(context.Background(), cfg, logger, nil, nil, nil)
		if err == nil {
			t.Error("expected error when protoflow config is nil")
		}
	})
}

func TestComposeEventMiddlewares(t *testing.T) {
	t.Run("with retry config", func(t *testing.T) {
		cfg := &protoflow.Config{
			RetryMaxRetries:      3,
			RetryInitialInterval: 100,
			RetryMaxInterval:     1000,
		}

		middlewares := composeEventMiddlewares(cfg)

		expectedCount := 8
		if len(middlewares) != expectedCount {
			t.Errorf("expected %d middlewares, got %d", expectedCount, len(middlewares))
		}
	})

	t.Run("with default config", func(t *testing.T) {
		cfg := &protoflow.Config{}
		middlewares := composeEventMiddlewares(cfg)

		if len(middlewares) == 0 {
			t.Error("expected at least one middleware")
		}
	})

	t.Run("with zero retry values", func(t *testing.T) {
		cfg := &protoflow.Config{
			RetryMaxRetries:      0,
			RetryInitialInterval: 0,
			RetryMaxInterval:     0,
		}

		middlewares := composeEventMiddlewares(cfg)
		if len(middlewares) != 8 {
			t.Errorf("expected 8 middlewares, got %d", len(middlewares))
		}
	})
}

func TestPoisonQueueFilter(t *testing.T) {
	filter := poisonQueueFilter()

	t.Run("nil error returns false", func(t *testing.T) {
		if filter(nil) {
			t.Error("expected filter to return false for nil error")
		}
	})

	t.Run("unprocessable error returns true", func(t *testing.T) {
		if !filter(protoflow.ErrUnprocessable) {
			t.Error("expected filter to return true for ErrUnprocessable")
		}
	})

	t.Run("unprocessable error pointer returns true", func(t *testing.T) {
		err := &protoflow.UnprocessableEventError{}
		if !filter(err) {
			t.Error("expected filter to return true for *UnprocessableEventError")
		}
	})

	t.Run("validation error returns true", func(t *testing.T) {
		validationErr := domain.ErrValidations{
			Errors: []string{"test: invalid"},
		}
		if !filter(validationErr) {
			t.Error("expected filter to return true for ErrValidations")
		}
	})

	t.Run("empty validation errors returns true", func(t *testing.T) {
		validationErr := domain.ErrValidations{Errors: []string{}}
		if !filter(validationErr) {
			t.Error("expected filter to return true for empty ErrValidations")
		}
	})

	t.Run("multiple validation errors returns true", func(t *testing.T) {
		validationErr := domain.ErrValidations{
			Errors: []string{"field1: required", "field2: invalid", "field3: too long"},
		}
		if !filter(validationErr) {
			t.Error("expected filter to return true for ErrValidations with multiple errors")
		}
	})

	t.Run("regular error returns false", func(t *testing.T) {
		if filter(errors.New("some error")) {
			t.Error("expected filter to return false for regular error")
		}
	})

	t.Run("wrapped validation error returns true", func(t *testing.T) {
		validationErr := domain.ErrValidations{Errors: []string{"field: required"}}
		wrappedErr := errors.Join(errors.New("validation failed"), validationErr)

		if !filter(wrappedErr) {
			t.Error("expected filter to return true for wrapped ErrValidations")
		}
	})

	t.Run("wrapped unprocessable error returns true", func(t *testing.T) {
		wrappedErr := errors.Join(errors.New("outer error"), protoflow.ErrUnprocessable)
		if !filter(wrappedErr) {
			t.Error("expected filter to return true for wrapped ErrUnprocessable")
		}
	})

	t.Run("deeply wrapped validation error returns true", func(t *testing.T) {
		validationErr := domain.ErrValidations{Errors: []string{"field: required"}}
		multiErr := errors.Join(
			errors.New("layer 1"),
			errors.New("layer 2"),
			validationErr,
		)

		if !filter(multiErr) {
			t.Error("expected filter to return true for deeply wrapped ErrValidations")
		}
	})
}

func TestStartEventService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil service returns immediately", func(t *testing.T) {
		done := make(chan struct{})
		go func() {
			StartEventService(context.Background(), nil, logger)
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("StartEventService should exit quickly with nil service")
		}
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})
		go func() {
			StartEventService(ctx, nil, logger)
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("StartEventService should exit when context is done")
		}
	})
}

func TestLogEventServiceStartup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	t.Run("nil service does not panic", func(t *testing.T) {
		// Should not panic when service is nil
		logEventServiceStartup(logger, nil)
	})
}
