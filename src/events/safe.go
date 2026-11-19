package events

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
)

type SafeOptions struct {
	Timeout time.Duration
	Logger  *slog.Logger
	Retry   func(error) bool
}

func Safe(handler gateway.Handler, options SafeOptions) gateway.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, event gateway.Event) (result error) {
		parent := ctx
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("event handler panicked", "event", event.Name, "panic", fmt.Sprint(recovered), "stack", string(debug.Stack()))
				result = nil
			}
		}()
		err := handler(ctx, event)
		if err == nil {
			return nil
		}
		if parent.Err() != nil || (options.Retry != nil && options.Retry(err)) {
			return err
		}
		logger.Error("event handler failed", "event", event.Name, "error", err)
		return nil
	}
}
