package events

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

type GoOptions struct {
	Timeout time.Duration
	Logger  *slog.Logger
	Name    string
}

func Go(ctx context.Context, options GoOptions, task func(context.Context) error) <-chan struct{} {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	name := options.Name
	if name == "" {
		name = "background task"
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		parent := ctx
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("background task panicked", "task", name, "panic", fmt.Sprint(recovered), "stack", string(debug.Stack()))
			}
		}()
		if err := task(ctx); err != nil && parent.Err() == nil {
			logger.Error("background task failed", "task", name, "error", err)
		}
	}()
	return done
}
