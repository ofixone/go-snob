package main

import "go.uber.org/zap/zapcore"

type Config struct {
	LogLevel zapcore.Level `long:"log-level" description:"Log level: panic, fatal, warn or warning, info, debug" env:"LOG_LEVEL" required:"true"`

	HTTPListenAddr string `long:"http-listen-addr" description:"Listening host:port for public http-server" env:"HTTP_LISTEN_ADDR" required:"true"`
}
