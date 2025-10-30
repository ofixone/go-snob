package pipeline

import (
	"encoding/json"
	"errors"
	"net/http"
)

func EncodeJSON[T any]() HandlerIn[T] {
	return func(ctx *Ctx, t T) error {
		ctx.Writer.Header().Add("Content-Type", "application/json; charset=utf-8")

		err := json.NewEncoder(ctx.Writer).Encode(t)
		if err != nil {
			http.Error(ctx.Writer, "bad response body decode", http.StatusInternalServerError)
		}
		return nil
	}
}

func AllowedMethods(methods ...string) MiddlewareFunc {
	allowed := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		allowed[m] = struct{}{}
	}

	return func(ctx *Ctx, next NextFunc) {
		if _, ok := allowed[ctx.Request.Method]; ok {
			next()
			return
		}

		http.Error(ctx.Writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func AllowedContentType(contentType ...string) MiddlewareFunc {
	allowed := make(map[string]struct{}, len(contentType))
	for _, m := range contentType {
		allowed[m] = struct{}{}
	}
	return func(ctx *Ctx, next NextFunc) {
		if _, ok := allowed[ctx.Request.Header.Get("Content-Type")]; ok {
			next()
			return
		}

		http.Error(ctx.Writer, "unsupported media type", http.StatusUnsupportedMediaType)
	}
}

func DecodeJSON[T any]() HandlerOut[T] {
	return func(ctx *Ctx) (t T, err error) {
		if err := json.NewDecoder(ctx.Request.Body).Decode(&t); err != nil {
			http.Error(ctx.Writer, "bad request body decode", http.StatusBadRequest)

			return t, errors.New("bad request")
		}
		return t, nil
	}
}
