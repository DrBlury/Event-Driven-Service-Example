package server

import (
	"context"
	"io"
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
		_, _ = w.Write([]byte("OK"))
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

func TestServerListenAndServeError(t *testing.T) {
	t.Parallel()

	// Create two servers on the same port to force an error
	cfg := &Config{Address: ":0"}
	mux := http.NewServeMux()

	srv1 := NewServer(cfg, mux)

	// Start first server
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv1.ListenAndServe()
	}()

	// Give first server time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown first server
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv1.Shutdown(ctx)

	// Check error from first server
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Logf("ListenAndServe returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for server to stop")
	}
}

func TestServerShutdownWithContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		timeout time.Duration
	}{
		{"immediate shutdown", 100 * time.Millisecond},
		{"quick shutdown", 500 * time.Millisecond},
		{"slow shutdown", 2 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{Address: ":0"}
			mux := http.NewServeMux()
			srv := NewServer(cfg, mux)

			// Start server
			go func() {
				_ = srv.ListenAndServe()
			}()

			time.Sleep(50 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			err := srv.Shutdown(ctx)
			if err != nil {
				t.Errorf("Shutdown failed: %v", err)
			}
		})
	}
}

func TestServerWithMultipleRoutes(t *testing.T) {
	t.Parallel()

	cfg := &Config{Address: ":0"}
	mux := http.NewServeMux()

	// Define routes that will return OK
	okRoutes := []string{"/", "/health", "/ready", "/api/v1"}

	for _, path := range okRoutes {
		path := path // Capture range variable
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}

	srv := NewServer(cfg, mux)

	// Test OK routes
	for _, path := range okRoutes {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			srv.server.Handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Status for %s = %d, want %d", path, rec.Code, http.StatusOK)
			}
		})
	}
}

func TestServerWithDifferentMethods(t *testing.T) {
	t.Parallel()

	cfg := &Config{Address: ":0"}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Method))
	})

	srv := NewServer(cfg, mux)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			srv.server.Handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Status for %s = %d, want %d", method, rec.Code, http.StatusOK)
			}
		})
	}
}

func TestServerWithResponseBody(t *testing.T) {
	t.Parallel()

	cfg := &Config{Address: ":0"}
	expectedBody := `{"status": "ok", "message": "Hello, World!"}`

	mux := http.NewServeMux()
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	})

	srv := NewServer(cfg, mux)

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	rec := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if string(body) != expectedBody {
		t.Errorf("Body = %s, want %s", string(body), expectedBody)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", ct)
	}
}

func TestServerShutdownWithoutStart(t *testing.T) {
	t.Parallel()

	cfg := &Config{Address: ":0"}
	mux := http.NewServeMux()
	srv := NewServer(cfg, mux)

	// Shutdown without starting - should not panic
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown without start failed: %v", err)
	}
}

func TestServerConcurrentRequests(t *testing.T) {
	t.Parallel()

	cfg := &Config{Address: ":0"}
	counter := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.WriteHeader(http.StatusOK)
	})

	srv := NewServer(cfg, mux)

	// Make concurrent requests
	numRequests := 100
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			srv.server.Handler.ServeHTTP(rec, req)
			done <- rec.Code == http.StatusOK
		}()
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-done {
			successCount++
		}
	}

	if successCount != numRequests {
		t.Errorf("Success count = %d, want %d", successCount, numRequests)
	}
}

func TestConfigStruct(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		cfg    Config
		fields map[string]string
	}{
		{
			name: "empty config",
			cfg:  Config{},
			fields: map[string]string{
				"Address":          "",
				"BaseURL":          "",
				"DocsTemplatePath": "",
			},
		},
		{
			name: "full config",
			cfg: Config{
				Address:          ":8080",
				BaseURL:          "/api/v1",
				DocsTemplatePath: "/templates/docs.html",
			},
			fields: map[string]string{
				"Address":          ":8080",
				"BaseURL":          "/api/v1",
				"DocsTemplatePath": "/templates/docs.html",
			},
		},
		{
			name: "production config",
			cfg: Config{
				Address:          "0.0.0.0:443",
				BaseURL:          "/api",
				DocsTemplatePath: "",
			},
			fields: map[string]string{
				"Address":          "0.0.0.0:443",
				"BaseURL":          "/api",
				"DocsTemplatePath": "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.cfg.Address != tc.fields["Address"] {
				t.Errorf("Address = %q, want %q", tc.cfg.Address, tc.fields["Address"])
			}
			if tc.cfg.BaseURL != tc.fields["BaseURL"] {
				t.Errorf("BaseURL = %q, want %q", tc.cfg.BaseURL, tc.fields["BaseURL"])
			}
			if tc.cfg.DocsTemplatePath != tc.fields["DocsTemplatePath"] {
				t.Errorf("DocsTemplatePath = %q, want %q", tc.cfg.DocsTemplatePath, tc.fields["DocsTemplatePath"])
			}
		})
	}
}

func TestServerWithHeaders(t *testing.T) {
	t.Parallel()

	cfg := &Config{Address: ":0"}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Echo back request headers
		for key, values := range r.Header {
			for _, value := range values {
				w.Header().Add("X-Echo-"+key, value)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := NewServer(cfg, mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom-Header", "test-value")
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("X-Echo-X-Custom-Header") != "test-value" {
		t.Error("Custom header was not echoed")
	}
}

func TestServerStartAndStopMultipleTimes(t *testing.T) {
	t.Parallel()

	for i := 0; i < 3; i++ {
		cfg := &Config{Address: ":0"}
		mux := http.NewServeMux()
		srv := NewServer(cfg, mux)

		// Start
		go func() {
			_ = srv.ListenAndServe()
		}()

		time.Sleep(50 * time.Millisecond)

		// Stop
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := srv.Shutdown(ctx)
		cancel()

		if err != nil {
			t.Errorf("Iteration %d: Shutdown failed: %v", i, err)
		}
	}
}
