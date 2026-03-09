package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	mu       sync.Mutex
	handlers []func(context.Context) error
)

func Add(handler any) {
	var wrapped func(context.Context) error

	switch h := handler.(type) {
	case func() error:
		wrapped = func(context.Context) error { return h() }
	case func():
		wrapped = func(context.Context) error {
			h()
			return nil
		}
	case func(context.Context) error:
		wrapped = h
	default:
		panic("closer.Add: unsupported handler type")
	}

	mu.Lock()
	defer mu.Unlock()
	handlers = append(handlers, wrapped)
}

func Close(ctx context.Context) error {
	mu.Lock()
	deferred := make([]func(context.Context) error, len(handlers))
	copy(deferred, handlers)
	mu.Unlock()

	var errs []error
	for i := len(deferred) - 1; i >= 0; i-- {
		if err := deferred[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("shutdown errors: %w", errors.Join(errs...))
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	handlers = nil
}
