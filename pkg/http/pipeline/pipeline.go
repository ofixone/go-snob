package pipeline

import (
	"context"
	"net/http"
	"reflect"
)

type Ctx struct {
	Writer  http.ResponseWriter
	Request *http.Request
	RCtx    context.Context
}

func NewCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		Writer:  w,
		Request: r,
		RCtx:    r.Context(),
	}
}

type NextFunc func()
type MiddlewareFunc func(ctx *Ctx, next NextFunc)

type Pipeline struct {
	middlewares []MiddlewareFunc
}

func NewPipeline(mw ...MiddlewareFunc) *Pipeline {
	return &Pipeline{middlewares: mw}
}

func (p *Pipeline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewCtx(w, r)
	var execIndex int

	var next NextFunc
	next = func() {
		if execIndex >= len(p.middlewares) {
			return
		}

		cur := p.middlewares[execIndex]
		execIndex++
		cur(ctx, next)
	}

	next()
}

func (p *Pipeline) WithMiddlewares(mw ...MiddlewareFunc) *Pipeline {
	p.middlewares = append(p.middlewares, mw...)
	return p
}

type HandlerInOut[In any, Out any] func(ctx *Ctx, in In) (Out, error)
type HandlerIn[In any] func(ctx *Ctx, in In) error
type HandlerOut[Out any] func(ctx *Ctx) (Out, error)

func In[I any](h HandlerIn[I]) MiddlewareFunc {
	return func(ctx *Ctx, next NextFunc) {
		if err := ctx.RCtx.Err(); err != nil {
			ctx.Writer.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		raw := ctx.RCtx.Value(reflect.TypeOf((*I)(nil)).Elem())
		if raw == nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		in := raw.(I)
		if err := h(ctx, in); err != nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		next()
	}
}

func Out[O any](h HandlerOut[O]) MiddlewareFunc {
	outType := reflect.TypeOf((*O)(nil)).Elem()

	return func(ctx *Ctx, next NextFunc) {
		if err := ctx.RCtx.Err(); err != nil {
			ctx.Writer.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		out, err := h(ctx)
		if err != nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx.RCtx = context.WithValue(ctx.RCtx, outType, out)
		next()
	}
}
func InOut[I any, O any](h HandlerInOut[I, O]) MiddlewareFunc {
	inType := reflect.TypeOf((*I)(nil)).Elem()
	outType := reflect.TypeOf((*O)(nil)).Elem()

	return func(ctx *Ctx, next NextFunc) {
		if err := ctx.RCtx.Err(); err != nil {
			ctx.Writer.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		raw := ctx.RCtx.Value(inType)
		if raw == nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		in := raw.(I)
		out, err := h(ctx, in)
		if err != nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx.RCtx = context.WithValue(ctx.RCtx, outType, out)
		next()
	}
}
