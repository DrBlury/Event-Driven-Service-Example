package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"drblury/event-driven-service/internal/observability"
)

type Server struct {
	server   *http.Server
	mu       sync.Mutex
	listener net.Listener
}

const READHEADERTIMEOUT = 5 * time.Second

func NewServer(cfg *Config, mux http.Handler) *Server {
	server := &http.Server{
		ReadHeaderTimeout: READHEADERTIMEOUT,
		Addr:              cfg.Address,
		Handler:           mux,
	}

	return &Server{
		server: server,
	}
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *Server) Start(errChan chan<- error) error {
	if s == nil || s.server == nil {
		return observability.Builder(context.Background(), "server.start", "server_not_configured").
			Public("server is not configured").
			New("server not configured")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return nil
	}

	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return observability.Builder(context.Background(), "server.start", "listen_failed").
			Public("server could not bind to its address").
			With("address", s.server.Addr).
			Wrap(err)
	}

	s.listener = listener
	go func() {
		err := s.server.Serve(listener)
		if errChan != nil {
			select {
			case errChan <- err:
			default:
			}
		}
	}()
	return nil
}

func (s *Server) Shutdown(ctxShutDown context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}

	err := s.server.Shutdown(ctxShutDown)
	s.mu.Lock()
	s.listener = nil
	s.mu.Unlock()
	if err == nil {
		return nil
	}

	return observability.Builder(ctxShutDown, "server.shutdown", "shutdown_failed").
		Public("server shutdown failed").
		Wrap(err)
}

func (s *Server) Address() string {
	if s == nil || s.server == nil {
		return ""
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.server.Addr
}
