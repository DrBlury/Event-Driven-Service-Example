package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	cfg := &Config{
		Address: ":8080",
		BaseURL: "/api",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := NewServer(cfg, mux)

	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
	if srv.server == nil {
		t.Error("server.server is nil")
	}
}

func TestNewServerWithDifferentConfigs(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"default port", ":80"},
		{"custom port", ":3000"},
		{"localhost", "localhost:8080"},
		{"all interfaces", "0.0.0.0:9000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Address: tt.address}
			srv := NewServer(cfg, http.NewServeMux())

			if srv == nil {
				t.Fatal("NewServer returned nil")
			}
		})
	}
}

func TestServerShutdown(t *testing.T) {
	cfg := &Config{Address: ":0"} // Use port 0 for random available port
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := NewServer(cfg, mux)

	// Start server in background
	go func() {
		_ = srv.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestREADHEADERTIMEOUT(t *testing.T) {
	if READHEADERTIMEOUT != 5*time.Second {
		t.Errorf("READHEADERTIMEOUT = %v, want %v", READHEADERTIMEOUT, 5*time.Second)
	}
}

func TestServerHandler(t *testing.T) {
	cfg := &Config{Address: ":0"}

	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := NewServer(cfg, mux)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Use the handler directly
	srv.server.Handler.ServeHTTP(rec, req)

	if !called {
		t.Error("Handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusOK)
	}
}
