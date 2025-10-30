package app

import (
	"context"
	"fmt"
	"go-snob/pkg/recoverer"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Module interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

type App struct {
	logger  *zap.Logger
	modules []Module
}

func NewApp(logger *zap.Logger) *App {
	return &App{logger: logger}
}

func (a *App) WithModules(modules ...Module) *App {
	a.modules = append(a.modules, modules...)
	return a
}

func (a *App) Run(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)
	for _, module := range a.modules {
		grp.Go(func() error {
			defer recoverer.Default()

			err := module.Run(ctx)
			if err != nil {
				return fmt.Errorf("can't init module: %w", err)
			}
			return nil
		})
	}

	grp.Go(func() error {
		defer recoverer.Default()

		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.logger.Info("shutting down app gracefully...")

		if err := a.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown error: %w", err)
		}
		a.logger.Info("shutdown complete")
		return nil
	})

	if err := grp.Wait(); err != nil {
		a.logger.Error("app exited with error", zap.Error(err))
	} else {
		a.logger.Info("app exited cleanly")
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	for _, module := range a.modules {
		// TODO: add the possibility to shut them down via goroutines. For now, I've left only 5 seconds for stopping,
		// which is probably not enough for the rest of the modules if one of them gets stuck.
		// TODO: add the possibility to notify this function about an error during module cleanup.
		// This could be useful if something goes wrong, so we can highlight it and terminate the app.
		err := module.Stop(ctx)
		if err != nil {
			return fmt.Errorf("can't cleanup module: %w", err)
		}
	}

	return nil
}
