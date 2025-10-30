package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"go-snob/pkg/http/pipeline"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"
)

func RawWebhookLog(logger *zap.Logger) pipeline.MiddlewareFunc {
	return func(ctx *pipeline.Ctx, next pipeline.NextFunc) {
		buf := new(bytes.Buffer)

		tee := io.TeeReader(ctx.Request.Body, buf)

		bodyBytes, err := io.ReadAll(tee)
		if err != nil {
			logger.Error("failed to read body", zap.Error(err))
			return
		}
		_ = ctx.Request.Body.Close()

		// TODO: need to restore bytes cause reader has read them till EOF, have to think about good solution
		ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		if f, err := os.CreateTemp(".", "webhook_*.json"); err == nil {
			_, _ = f.Write(buf.Bytes())
			_ = f.Close()
			logger.Info("raw webhook was logged", zap.String("path", f.Name()))
		}

		next()
	}
}

func CheckSecret(secret string) pipeline.MiddlewareFunc {
	return func(ctx *pipeline.Ctx, next pipeline.NextFunc) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil || len(body) == 0 {
			http.Error(ctx.Writer, "empty payload", http.StatusBadRequest)
			return
		}

		headerSig := ctx.Request.Header.Get("X-Gitea-Signature")
		if headerSig == "" {
			http.Error(ctx.Writer, "signature header missing", http.StatusBadRequest)
			return
		}

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		if headerSig != expectedSig {
			http.Error(ctx.Writer, "invalid hmac signature", http.StatusUnauthorized)
			return
		}

		// TODO: need to restore bytes cause reader has read them till EOF, have to think about good solution
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

		next()
	}
}

func Push[P any](ch chan<- P) pipeline.HandlerIn[P] {
	return func(ctx *pipeline.Ctx, in P) error {
		select {
		case <-ctx.Request.Context().Done():
			http.Error(ctx.Writer, "request cancelled", http.StatusRequestTimeout)
		case ch <- in:
		}

		return nil
	}
}
