package recoverer

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()

	cfg.Level = zap.NewAtomicLevelAt(zap.PanicLevel)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	l := zap.Must(cfg.Build())
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)

	logger = l
}

var Default = NewRecoverer(func(v interface{}) {
	logger.Error(
		"recovered from panic",
		zap.Error(
			errors.New(
				strings.ReplaceAll(
					strings.ReplaceAll(fmt.Sprintf("%v", v), `\`, `\\`), `"`, `\"`,
				),
			),
		),
	)
})

func NewRecoverer(processor func(v interface{})) func() {
	return func() {
		if r := recover(); r != nil {
			processor(r)
		}
	}
}

func SetLogger(logr *zap.Logger) {
	logger = logr
}
