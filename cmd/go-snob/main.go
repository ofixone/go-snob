package main

import (
	"context"
	"fmt"
	"go-snob/pkg/app"
	"go-snob/pkg/recoverer"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Version = "0.1.0"
)

func main() {
	cfg := newCfg()
	logger := newLogger(cfg.LogLevel)
	recoverer.SetLogger(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	logger.Info("starting app init..")
	err := app.NewApp(logger).WithModules(
		NewHttpServer(logger, cfg.HTTPListenAddr),
	).Run(ctx)
	if err != nil {
		logger.Fatal("app run failed", zap.Error(err))
	}
}

func fatalJSONLog(msg string, err error) string {
	escape := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
	}
	errString := ""
	if err != nil {
		errString = err.Error()
	}
	return fmt.Sprintf(
		`{"level":"fatal","ts":"%s","msg":"%s","error":"%s"}`,
		time.Now().Format(time.RFC3339),
		escape(msg),
		escape(errString),
	)
}

func newCfg() Config {
	var cfg Config
	parser := flags.NewParser(&cfg, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(fatalJSONLog("Failed to parse config.", err))
	}

	return cfg
}

func newLogger(lvl zapcore.Level) *zap.Logger {
	cfg := zap.NewProductionConfig()

	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]interface{}{
		"version": Version,
	}
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger := zap.Must(cfg.Build())
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	return logger
}
