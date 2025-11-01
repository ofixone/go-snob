package main

import (
	"context"
	"fmt"
	apihttpwebhook "go-snob/cmd/go-snob/api/http/webhook"
	"go-snob/internal"
	"go-snob/internal/actor/ai"
	"go-snob/internal/actor/vcs/gitea"
	"go-snob/pkg/app"
	"go-snob/pkg/giteawebhook"
	"go-snob/pkg/http"
	"go-snob/pkg/recoverer"
	"go-snob/pkg/restyprometheus"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"resty.dev/v3"
)

// TODO: need to go through whole code base and fix naming sometimes :)

const (
	Version = "0.1.0"
)

func main() {
	cfg := newCfg()
	yamlCfg := newYamlCfg(cfg.PathToYAMLCfg)
	logger := newLogger(cfg.LogLevel)
	recoverer.SetLogger(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	giteaClient := gitea.NewClient(
		restyprometheus.NewClient(resty.New(), "go-snob", "gitea_client"),
		"http://localhost:3000/api/v1",
		cfg.SNOBUserGiteaToken,
	)
	aiClient := ai.NewClient(
		restyprometheus.NewClient(resty.New(), "go-snob", "ai_client"),
		logger, "https://foundation-models.api.cloud.ru/v1/chat/completions",
		cfg.CloudRuFoundationalModelsKey,
		yamlCfg.SystemPrompt,
	)
	orch := internal.NewOrchestrator(aiClient, giteaClient, logger)

	webhook := giteawebhook.NewWebhook(orch.Handler, logger).WithDebug().WithSecret(cfg.WebhookGiteaSecret)

	server := apihttpwebhook.NewServer(webhook)

	logger.Info("starting app init..")
	err := app.NewApp(logger).WithModules(
		http.NewServer(logger, cfg.HTTPListenAddr).
			WithPingHandler().
			WithHandler("/webhook", server.GiteaWebhook()),
		webhook,
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

func newYamlCfg(path string) YAMLConfig {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(fatalJSONLog("Failed to get working directory.", err))
	}
	b, err := os.ReadFile(filepath.Join(wd, path))
	if err != nil {
		log.Fatal(fatalJSONLog("Failed to read YAML config.", err))
	}
	var cfg YAMLConfig
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal(fatalJSONLog("Failed to unmarshal YAML config.", err))
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
