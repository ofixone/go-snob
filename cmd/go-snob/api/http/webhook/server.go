package webhook

import (
	"go-snob/pkg/giteawebhook"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	logger  *zap.Logger
	webhook *giteawebhook.Webhook
}

func NewServer(webhook *giteawebhook.Webhook) *Server {
	return &Server{webhook: webhook}
}

func (s *Server) GiteaWebhook() http.Handler {
	return s.webhook.Handler()
}
