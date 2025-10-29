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
			ctx.Writer.WriteHeader(http.StatusBadRequest)
		}
		return nil
	}
}

func AllowMethods(methods ...string) MiddlewareFunc {
	allowed := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		allowed[m] = struct{}{}
	}

	return func(ctx *Ctx, next NextFunc) {
		if _, ok := allowed[ctx.Request.Method]; ok {
			next()
			return
		}

		ctx.Writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func DecodeJSON[T any]() HandlerOut[T] {
	return func(ctx *Ctx) (t T, err error) {
		if err := json.NewDecoder(ctx.Request.Body).Decode(&t); err != nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return t, errors.New("bad request")
		}
		return t, nil
	}
}
