package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"secure-rag-platform/services/ai-inference/internal/closer"
)

type App struct {
	components  []func() error
	stopTimeout time.Duration
}

func New() *App {
	closer.Reset()
	return &App{stopTimeout: 10 * time.Second}
}

func (a *App) Add(component func() error) {
	a.components = append(a.components, component)
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, len(a.components))
	var wg sync.WaitGroup

	for _, component := range a.components {
		wg.Add(1)
		go func(c func() error) {
			defer wg.Done()
			if err := c(); err != nil {
				errCh <- err
			}
		}(component)
	}

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errCh:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.stopTimeout)
	defer cancel()

	if err := closer.Close(shutdownCtx); err != nil && runErr == nil {
		runErr = err
	}

	wg.Wait()
	return runErr
}
