package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/internal/usecase"

	"github.com/drblury/apiweaver/router"
	"go.uber.org/fx"
)

func TestBuildHTTPServer(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

	appLogic, err := usecase.NewAppLogic(nil, logger, nil, nil)
	if err != nil {
		t.Fatalf("NewAppLogic returned error: %v", err)
	}

	t.Run("missing server config", func(t *testing.T) {
		t.Parallel()
		_, err := buildHTTPServer(&domain.Info{}, &router.Config{}, nil, appLogic, logger)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("missing router config", func(t *testing.T) {
		t.Parallel()
		_, err := buildHTTPServer(&domain.Info{}, nil, &server.Config{Address: "127.0.0.1:0"}, appLogic, logger)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()
		srv, err := buildHTTPServer(
			&domain.Info{Version: "1.0.0"},
			&router.Config{Timeout: 30 * time.Second},
			&server.Config{Address: "127.0.0.1:0", BaseURL: ""},
			appLogic,
			logger,
		)
		if err != nil {
			t.Fatalf("buildHTTPServer returned error: %v", err)
		}
		if srv == nil {
			t.Fatal("expected server")
		}
	})
}

func TestRegisterHTTPServerLifecycleStartsAndStopsServer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	appLogic, err := usecase.NewAppLogic(nil, logger, nil, nil)
	if err != nil {
		t.Fatalf("NewAppLogic returned error: %v", err)
	}

	serverCfg := &server.Config{Address: "127.0.0.1:0"}
	srv, err := buildHTTPServer(
		&domain.Info{Version: "1.0.0"},
		&router.Config{Timeout: 30 * time.Second},
		serverCfg,
		appLogic,
		logger,
	)
	if err != nil {
		t.Fatalf("buildHTTPServer returned error: %v", err)
	}

	app := fx.New(
		fx.Supply(srv),
		fx.Supply(serverCfg),
		fx.Supply(logger),
		fx.Invoke(registerHTTPServerLifecycle),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx.New returned error: %v", err)
	}

	startCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("app.Start returned error: %v", err)
	}

	addr := srv.Address()
	client := http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + addr + "/info/openapi.json")
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		t.Fatalf("app.Stop returned error: %v", err)
	}

	resp, err = client.Get("http://" + addr + "/info/openapi.json")
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected request failure after shutdown")
	}
}
