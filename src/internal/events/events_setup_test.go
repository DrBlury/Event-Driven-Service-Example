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

	t.Run("with high retry values", func(t *testing.T) {
		cfg := &protoflow.Config{
			RetryMaxRetries:      100,
			RetryInitialInterval: 10000,
			RetryMaxInterval:     100000,
		}

		middlewares := composeEventMiddlewares(cfg)
		if len(middlewares) != 8 {
			t.Errorf("expected 8 middlewares, got %d", len(middlewares))
		}
	})

	t.Run("middleware order is consistent", func(t *testing.T) {
		cfg := &protoflow.Config{
			RetryMaxRetries:      3,
			RetryInitialInterval: 100,
			RetryMaxInterval:     1000,
		}

		middlewares1 := composeEventMiddlewares(cfg)
		middlewares2 := composeEventMiddlewares(cfg)

		if len(middlewares1) != len(middlewares2) {
			t.Error("middleware count should be consistent")
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

	t.Run("unprocessable errors", func(t *testing.T) {
		testPoisonQueueFilterUnprocessable(t, filter)
	})

	t.Run("validation errors", func(t *testing.T) {
		testPoisonQueueFilterValidation(t, filter)
	})

	t.Run("wrapped errors", func(t *testing.T) {
		testPoisonQueueFilterWrapped(t, filter)
	})

	t.Run("regular errors return false", func(t *testing.T) {
		regularErrors := []error{
			errors.New("connection refused"),
			errors.New("timeout"),
			errors.New("internal error"),
			context.Canceled,
			context.DeadlineExceeded,
		}

		for _, err := range regularErrors {
			if filter(err) {
				t.Errorf("expected filter to return false for %v", err)
			}
		}
	})
}

func testPoisonQueueFilterUnprocessable(t *testing.T, filter func(error) bool) {
	t.Helper()

	if !filter(protoflow.ErrUnprocessable) {
		t.Error("expected filter to return true for ErrUnprocessable")
	}

	err := &protoflow.UnprocessableEventError{}
	if !filter(err) {
		t.Error("expected filter to return true for *UnprocessableEventError")
	}
}

func testPoisonQueueFilterValidation(t *testing.T, filter func(error) bool) {
	t.Helper()

	validationErr := domain.ErrValidations{Errors: []string{"test: invalid"}}
	if !filter(validationErr) {
		t.Error("expected filter to return true for ErrValidations")
	}

	emptyValidationErr := domain.ErrValidations{Errors: []string{}}
	if !filter(emptyValidationErr) {
		t.Error("expected filter to return true for empty ErrValidations")
	}

	multiValidationErr := domain.ErrValidations{
		Errors: []string{"field1: required", "field2: invalid", "field3: too long"},
	}
	if !filter(multiValidationErr) {
		t.Error("expected filter to return true for ErrValidations with multiple errors")
	}

	if filter(errors.New("some error")) {
		t.Error("expected filter to return false for regular error")
	}
}

func testPoisonQueueFilterWrapped(t *testing.T, filter func(error) bool) {
	t.Helper()

	validationErr := domain.ErrValidations{Errors: []string{"field: required"}}
	wrappedErr := errors.Join(errors.New("validation failed"), validationErr)
	if !filter(wrappedErr) {
		t.Error("expected filter to return true for wrapped ErrValidations")
	}

	wrappedUnprocessable := errors.Join(errors.New("outer error"), protoflow.ErrUnprocessable)
	if !filter(wrappedUnprocessable) {
		t.Error("expected filter to return true for wrapped ErrUnprocessable")
	}

	deeplyWrapped := errors.Join(
		errors.New("layer 1"),
		errors.New("layer 2"),
		validationErr,
	)
	if !filter(deeplyWrapped) {
		t.Error("expected filter to return true for deeply wrapped ErrValidations")
	}
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

	t.Run("nil logger does not panic", func(t *testing.T) {
		done := make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("StartEventService panicked: %v", r)
				}
				close(done)
			}()
			StartEventService(context.Background(), nil, nil)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("StartEventService should exit quickly")
		}
	})
}

func TestLogEventServiceStartup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	t.Run("nil service does not panic", func(t *testing.T) {
		// Should not panic when service is nil
		logEventServiceStartup(logger, nil)
	})

	t.Run("nil logger does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("logEventServiceStartup panicked with nil logger: %v", r)
			}
		}()
		logEventServiceStartup(nil, nil)
	})
}

func TestBuildEventServiceWithVariousConfigs(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Only test nil config cases that return errors early without requiring transport setup
	testCases := []struct {
		name         string
		cfg          *Config
		protoflowCfg *protoflow.Config
		expectError  bool
	}{
		{
			name:         "both configs nil",
			cfg:          nil,
			protoflowCfg: nil,
			expectError:  true,
		},
		{
			name:         "cfg nil, protoflow not nil",
			cfg:          nil,
			protoflowCfg: &protoflow.Config{},
			expectError:  true,
		},
		{
			name:         "cfg not nil, protoflow nil",
			cfg:          &Config{},
			protoflowCfg: nil,
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := BuildEventService(context.Background(), tc.cfg, logger, nil, nil, tc.protoflowCfg)
			if tc.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfigStruct(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DemoConsumeQueue:    "demo-consume",
		DemoPublishQueue:    "demo-publish",
		ExampleConsumeQueue: "example-consume",
		ExamplePublishQueue: "example-publish",
	}

	if cfg.DemoConsumeQueue != "demo-consume" {
		t.Errorf("DemoConsumeQueue = %q, want 'demo-consume'", cfg.DemoConsumeQueue)
	}
	if cfg.DemoPublishQueue != "demo-publish" {
		t.Errorf("DemoPublishQueue = %q, want 'demo-publish'", cfg.DemoPublishQueue)
	}
	if cfg.ExampleConsumeQueue != "example-consume" {
		t.Errorf("ExampleConsumeQueue = %q, want 'example-consume'", cfg.ExampleConsumeQueue)
	}
	if cfg.ExamplePublishQueue != "example-publish" {
		t.Errorf("ExamplePublishQueue = %q, want 'example-publish'", cfg.ExamplePublishQueue)
	}
}

func TestPoisonQueueFilterEdgeCases(t *testing.T) {
	t.Parallel()

	filter := poisonQueueFilter()

	t.Run("UnprocessableEventError with message", func(t *testing.T) {
		err := &protoflow.UnprocessableEventError{}
		if !filter(err) {
			t.Error("should return true for UnprocessableEventError")
		}
	})

	t.Run("ErrValidations with special characters", func(t *testing.T) {
		err := domain.ErrValidations{
			Errors: []string{
				"field: contains 'quotes'",
				`field: contains "double quotes"`,
				"field: contains <>&",
			},
		}
		if !filter(err) {
			t.Error("should return true for ErrValidations with special chars")
		}
	})

	t.Run("ErrValidations with empty strings", func(t *testing.T) {
		err := domain.ErrValidations{
			Errors: []string{"", "", ""},
		}
		if !filter(err) {
			t.Error("should return true for ErrValidations with empty strings")
		}
	})

	t.Run("multiple wrapped errors", func(t *testing.T) {
		innerErr := domain.ErrValidations{Errors: []string{"inner"}}
		midErr := errors.Join(errors.New("mid"), innerErr)
		outerErr := errors.Join(errors.New("outer"), midErr)

		if !filter(outerErr) {
			t.Error("should return true for multiply wrapped validation error")
		}
	})
}

func TestComposeEventMiddlewaresVariousConfigs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		retryMaxRetries int
		retryInitial    time.Duration
		retryMax        time.Duration
		expectedCount   int
	}{
		{
			name:            "zero retries",
			retryMaxRetries: 0,
			retryInitial:    0,
			retryMax:        0,
			expectedCount:   8,
		},
		{
			name:            "small retries",
			retryMaxRetries: 1,
			retryInitial:    10 * time.Millisecond,
			retryMax:        100 * time.Millisecond,
			expectedCount:   8,
		},
		{
			name:            "large retries",
			retryMaxRetries: 10,
			retryInitial:    1 * time.Second,
			retryMax:        10 * time.Second,
			expectedCount:   8,
		},
		{
			name:            "negative values (should be handled)",
			retryMaxRetries: -1,
			retryInitial:    -100 * time.Millisecond,
			retryMax:        -1 * time.Second,
			expectedCount:   8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := &protoflow.Config{
				RetryMaxRetries:      tc.retryMaxRetries,
				RetryInitialInterval: tc.retryInitial,
				RetryMaxInterval:     tc.retryMax,
			}

			middlewares := composeEventMiddlewares(cfg)
			if len(middlewares) != tc.expectedCount {
				t.Errorf("expected %d middlewares, got %d", tc.expectedCount, len(middlewares))
			}
		})
	}
}
