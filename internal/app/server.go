package app

import (
	"context"
	"net/http"
	"time"

	"ecommerce-ai-assistant/internal/platform/logging"
)

type Server struct {
	server *http.Server
	logger *logging.Logger
}

func NewServer(addr string, handler http.Handler, logger *logging.Logger) *Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return &Server{server: srv, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("server starting", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	return s.server.Shutdown(ctx)
}
