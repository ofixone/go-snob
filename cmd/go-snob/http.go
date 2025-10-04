package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type HTTPServer struct {
	logger *zap.Logger
	server *http.Server
}

func NewHttpServer(logger *zap.Logger, addr string) *HTTPServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &HTTPServer{
		logger: logger,
		server: srv,
	}
}

func (s *HTTPServer) Run(_ context.Context) error {
	s.logger.Info("starting HTTP server", zap.String("addr", s.server.Addr))

	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("error on listening and serving HTTP server", zap.Error(err))
		return err
	}
	s.logger.Info("HTTP server closed")
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("start shutdown HTTP server", zap.String("addr", s.server.Addr))
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("can't shutdown HTTP server: %w", err)
	}
	s.logger.Info("HTTP server shutdown gracefully")

	return nil
}
