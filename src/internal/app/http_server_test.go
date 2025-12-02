package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/internal/usecase"

	"github.com/drblury/apiweaver/router"
)

func TestRunHTTPServer(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	t.Run("nil server", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Server: &server.Config{Address: ":0"}}
		errChan := make(chan error, 1)
		runHTTPServer(nil, cfg, logger, errChan)
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		srv := &server.Server{}
		errChan := make(chan error, 1)
		runHTTPServer(srv, nil, logger, errChan)
	})

	t.Run("nil server config", func(t *testing.T) {
		t.Parallel()
		srv := &server.Server{}
		cfg := &Config{Server: nil}
		errChan := make(chan error, 1)
		runHTTPServer(srv, cfg, logger, errChan)
	})

	t.Run("with nil error channel", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Server: &server.Config{Address: ":0"}}
		runHTTPServer(nil, cfg, logger, nil)
	})

	t.Run("all nils does not panic", func(t *testing.T) {
		t.Parallel()
		runHTTPServer(nil, nil, nil, nil)
	})
}

func TestMonitorHTTPServerErrors(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil channel does not panic", func(t *testing.T) {
		t.Parallel()
		monitorHTTPServerErrors(context.Background(), nil, logger)
	})

	t.Run("cancelled context", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, 1)
		monitorHTTPServerErrors(ctx, errChan, logger)
		cancel()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("context done", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, 1)
		monitorHTTPServerErrors(ctx, errChan, logger)
		cancel()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("various error types", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name string
			err  error
		}{
			{"nil error", nil},
			{"server closed", http.ErrServerClosed},
			{"random error", errors.New("some error")},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				errChan := make(chan error, 1)
				monitorHTTPServerErrors(context.Background(), errChan, logger)
				errChan <- tc.err
				time.Sleep(10 * time.Millisecond)
			})
		}
	})
}

func TestShutdownHTTPServer(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil server", func(t *testing.T) {
		t.Parallel()
		err := shutdownHTTPServer(nil, logger)
		if err != nil {
			t.Errorf("shutdownHTTPServer with nil server should not error: %v", err)
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		err := shutdownHTTPServer(nil, nil)
		if err != nil {
			t.Errorf("shutdownHTTPServer should not error for nil server: %v", err)
		}
	})
}

func TestBuildHTTPServer(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("nil config panics or errors", func(t *testing.T) {
		t.Parallel()
		defer func() {
			_ = recover() // Expected to recover from panic
		}()
		_, _ = buildHTTPServer(nil, nil, logger)
	})

	t.Run("various configs", func(t *testing.T) {
		t.Parallel()
		testBuildHTTPServerConfigs(t, logger)
	})
}

func testBuildHTTPServerConfigs(t *testing.T, logger *slog.Logger) {
	t.Helper()
	testCases := buildHTTPServerTestCases()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appLogic, _ := usecase.NewAppLogic(nil, logger)
			srv, err := buildHTTPServer(tc.cfg, appLogic, logger)
			if err != nil {
				t.Fatalf("buildHTTPServer failed for %s: %v", tc.name, err)
			}
			if srv == nil {
				t.Errorf("buildHTTPServer returned nil for %s", tc.name)
			}
		})
	}
}

func buildHTTPServerTestCases() []struct {
	name string
	cfg  *Config
} {
	return []struct {
		name string
		cfg  *Config
	}{
		{"minimal config", &Config{Server: &server.Config{Address: ":0"}, Router: &router.Config{}, Info: &domain.Info{}}},
		{"with base URL", &Config{Server: &server.Config{Address: ":0", BaseURL: "/api"}, Router: &router.Config{Timeout: 30 * time.Second}, Info: &domain.Info{Version: "2.0.0"}}},
		{"with CORS", &Config{Server: &server.Config{Address: ":0"}, Router: &router.Config{CORS: router.CORSConfig{Origins: []string{"*"}, Methods: []string{"GET", "POST"}}}, Info: &domain.Info{Version: "3.0.0"}}},
		{"full config", buildFullConfigTestCase()},
	}
}

func buildFullConfigTestCase() *Config {
	return &Config{
		Server: &server.Config{Address: ":8080", BaseURL: "/api/v1", DocsTemplatePath: ""},
		Router: &router.Config{
			Timeout:         60 * time.Second,
			CORS:            router.CORSConfig{AllowCredentials: true, Headers: []string{"*"}, Methods: []string{"GET", "POST"}, Origins: []string{"*"}},
			QuietdownRoutes: []string{"/healthz"},
			HideHeaders:     []string{"Authorization"},
		},
		Info: &domain.Info{Version: "2.0.0", BuildDate: "2024-06-15", Details: "Test", CommitHash: "abc123"},
	}
}

func TestHTTPServerLifecycle(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("shutdown without starting", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Server: &server.Config{Address: ":0", BaseURL: ""},
			Router: &router.Config{Timeout: 30 * time.Second},
			Info:   &domain.Info{Version: "1.0.0"},
		}

		appLogic, _ := usecase.NewAppLogic(nil, logger)
		srv, err := buildHTTPServer(cfg, appLogic, logger)
		if err != nil {
			t.Fatalf("buildHTTPServer failed: %v", err)
		}

		err = shutdownHTTPServer(srv, logger)
		if err != nil {
			t.Errorf("shutdownHTTPServer failed: %v", err)
		}
	})

	t.Run("start and shutdown", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Server: &server.Config{Address: ":0", BaseURL: ""},
			Router: &router.Config{Timeout: 30 * time.Second},
			Info:   &domain.Info{Version: "1.0.0"},
		}

		appLogic, _ := usecase.NewAppLogic(nil, logger)
		srv, err := buildHTTPServer(cfg, appLogic, logger)
		if err != nil {
			t.Fatalf("buildHTTPServer failed: %v", err)
		}

		errChan := make(chan error, 1)
		runHTTPServer(srv, cfg, logger, errChan)

		time.Sleep(50 * time.Millisecond)

		err = shutdownHTTPServer(srv, logger)
		if err != nil {
			t.Errorf("shutdownHTTPServer failed: %v", err)
		}
	})
}
