package events

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
)

func safeLogger(buffer *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buffer, nil))
}

func TestSafeSwallowsAndLogsErrors(t *testing.T) {
	var logs bytes.Buffer
	handler := Safe(func(context.Context, gateway.Event) error { return errors.New("403 missing access") }, SafeOptions{Logger: safeLogger(&logs)})
	if err := handler(context.Background(), gateway.Event{Name: "MESSAGE_CREATE"}); err != nil {
		t.Fatalf("handler error was returned: %v", err)
	}
	if !strings.Contains(logs.String(), "403 missing access") || !strings.Contains(logs.String(), "MESSAGE_CREATE") {
		t.Fatalf("error was not logged: %q", logs.String())
	}
}

func TestSafeRetryReturnsSelectedErrors(t *testing.T) {
	retry := errors.New("database unavailable")
	skip := errors.New("403 missing access")
	var logs bytes.Buffer
	for _, tc := range []struct {
		err  error
		want error
	}{{retry, retry}, {skip, nil}} {
		handler := Safe(func(context.Context, gateway.Event) error { return tc.err }, SafeOptions{
			Logger: safeLogger(&logs),
			Retry:  func(err error) bool { return errors.Is(err, retry) },
		})
		if got := handler(context.Background(), gateway.Event{}); !errors.Is(got, tc.want) && got != tc.want {
			t.Fatalf("Safe returned %v, want %v", got, tc.want)
		}
	}
}

func TestSafeRecoversPanics(t *testing.T) {
	var logs bytes.Buffer
	handler := Safe(func(context.Context, gateway.Event) error {
		var event *gateway.Event
		_ = event.Name
		return nil
	}, SafeOptions{Logger: safeLogger(&logs)})
	if err := handler(context.Background(), gateway.Event{Name: "INTERACTION_CREATE"}); err != nil {
		t.Fatalf("panic leaked as error: %v", err)
	}
	if !strings.Contains(logs.String(), "panicked") || !strings.Contains(logs.String(), "INTERACTION_CREATE") {
		t.Fatalf("panic was not logged: %q", logs.String())
	}
}

func TestSafeTimeoutCancelsContext(t *testing.T) {
	var logs bytes.Buffer
	handler := Safe(func(ctx context.Context, _ gateway.Event) error {
		<-ctx.Done()
		return ctx.Err()
	}, SafeOptions{Timeout: 10 * time.Millisecond, Logger: safeLogger(&logs)})
	if err := handler(context.Background(), gateway.Event{}); err != nil {
		t.Fatalf("timeout leaked as error: %v", err)
	}
	if !strings.Contains(logs.String(), "deadline exceeded") {
		t.Fatalf("timeout was not logged: %q", logs.String())
	}
}

func TestSafeKeepsShutdownError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	handler := Safe(func(ctx context.Context, _ gateway.Event) error { return ctx.Err() }, SafeOptions{Logger: safeLogger(&bytes.Buffer{})})
	if err := handler(ctx, gateway.Event{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("shutdown error was swallowed: %v", err)
	}
}
