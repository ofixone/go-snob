package main

import "go.uber.org/zap/zapcore"

type Config struct {
	LogLevel zapcore.Level `long:"log-level" description:"Log level: panic, fatal, warn or warning, info, debug" env:"LOG_LEVEL" required:"true"`

	HTTPListenAddr string `long:"http-listen-addr" description:"Listening host:port for public http-server" env:"HTTP_LISTEN_ADDR" required:"true"`

	// SNOBUserGiteaToken SNOB user gitea. Used for calling api by himself
	SNOBUserGiteaToken string `long:"snob-user-gitea-token" description:"User SNOB Gitea Token" env:"SNOB_USER_GITEA_TOKEN" required:"true"`

	// WebhookGiteaSecret webhook secret for HMAC signature
	WebhookGiteaSecret string `long:"webhook-gitea-secret" description:"Webhook Gitea Secret" env:"WEBHOOK_GITEA_SECRET" required:"true"`
}
