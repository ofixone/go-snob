package http

import (
	"context"
	"errors"
	"fmt"
	"go-snob/pkg/http/pipeline"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	server *http.Server
	mux    *http.ServeMux
}

func NewServer(logger *zap.Logger, addr string) *Server {
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{
		logger: logger,
		mux:    mux,
		server: srv,
	}
}

func (s *Server) WithPingHandler() *Server {
	return s.WithHandler("/ping", pipeline.NewPipeline(
		pipeline.AllowedMethods(http.MethodGet),
	))
}

func (s *Server) WithHandler(path string, handler http.Handler) *Server {
	s.mux.Handle(path, handler)
	return s
}

func (s *Server) WithHandlers(m map[string]http.Handler) *Server {
	for pattern, handler := range m {
		s.mux.Handle(pattern, handler)
	}

	return s
}

func (s *Server) Run(_ context.Context) error {
	s.logger.Info("starting HTTP server", zap.String("addr", s.server.Addr))

	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("error on listening and serving HTTP server", zap.Error(err))
		return err
	}
	s.logger.Info("HTTP server closed")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("start shutdown HTTP server", zap.String("addr", s.server.Addr))
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("can't shutdown HTTP server: %w", err)
	}
	s.logger.Info("HTTP server shutdown gracefully")

	return nil
}
