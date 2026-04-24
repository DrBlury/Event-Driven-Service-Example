package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
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
		return errors.New("server not configured")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return nil
	}

	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
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
	return err
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
