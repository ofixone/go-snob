package giteawebhook

import (
	"context"
	"go-snob/pkg/giteawebhook/middleware"
	"go-snob/pkg/http/pipeline"
	"go-snob/pkg/workerpool"
	"net/http"

	"go.uber.org/zap"
)

const defaultPayloadChanCapacity = 1024

type Webhook struct {
	logger *zap.Logger

	debug  bool
	secret string

	workers *workerpool.Pool[Payload]
}

func (wh *Webhook) Run(ctx context.Context) error {
	wh.logger.Info("start handling webhook..")
	wh.workers.Start(ctx)
	wh.logger.Info("successfully starting background webhook handling")
	return nil
}

func (wh *Webhook) Stop(ctx context.Context) error {
	wh.logger.Info("stop handling webhook..")
	wh.workers.Shutdown(ctx)
	wh.logger.Info("webhook handling stopped")
	return nil
}

func NewWebhook(proc workerpool.Processor[Payload], logger *zap.Logger) *Webhook {
	return &Webhook{
		logger:  logger,
		workers: workerpool.NewPool[Payload](proc, 5, defaultPayloadChanCapacity),
	}
}

func (wh *Webhook) WithDebug() *Webhook {
	wh.debug = true
	return wh
}

func (wh *Webhook) WithSecret(secret string) *Webhook {
	wh.secret = secret
	return wh
}

func (wh *Webhook) WithPayloadChanCapacity(capacity int) *Webhook {
	wh.workers.SetChan(make(chan Payload, capacity))
	return wh
}

func (wh *Webhook) Handler() http.Handler {
	p := pipeline.NewPipeline()
	if wh.debug {
		p.WithMiddlewares(middleware.RawWebhookLog(wh.logger))
	}

	p.WithMiddlewares(
		pipeline.AllowedMethods(http.MethodPost),
		pipeline.AllowedContentType("application/json"),
	)

	if wh.secret != "" {
		p.WithMiddlewares(middleware.CheckSecret(wh.secret))
	}

	p.WithMiddlewares(
		pipeline.Out(pipeline.DecodeJSON[Payload]()),
		pipeline.In(middleware.Push[Payload](wh.workers.WrChan())),
	)

	return p
}
