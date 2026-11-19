package events

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

type lockedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

func waitForLog(t *testing.T, logs *lockedBuffer, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(logs.String(), want) {
		if time.Now().After(deadline) {
			t.Fatalf("log does not contain %q: %q", want, logs.String())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestGoReturnsBeforeTheTaskFinishes(t *testing.T) {
	release := make(chan struct{})
	done := Go(context.Background(), GoOptions{Logger: slog.New(slog.DiscardHandler)}, func(context.Context) error {
		<-release
		return nil
	})
	select {
	case <-done:
		t.Fatal("Go waited for the task")
	default:
	}
	close(release)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("task did not run")
	}
}

func TestGoRecoversPanics(t *testing.T) {
	logs := &lockedBuffer{}
	Go(context.Background(), GoOptions{Name: "search", Logger: slog.New(slog.NewTextHandler(logs, nil))}, func(context.Context) error {
		var values map[string]int
		values["a"] = 1
		return nil
	})
	waitForLog(t, logs, "background task panicked")
	if !strings.Contains(logs.String(), "search") {
		t.Fatalf("task name missing: %q", logs.String())
	}
}

func TestGoLogsErrors(t *testing.T) {
	logs := &lockedBuffer{}
	Go(context.Background(), GoOptions{Logger: slog.New(slog.NewTextHandler(logs, nil))}, func(context.Context) error {
		return errors.New("search failed")
	})
	waitForLog(t, logs, "search failed")
}

func TestGoTimeoutCancelsTask(t *testing.T) {
	logs := &lockedBuffer{}
	Go(context.Background(), GoOptions{Timeout: 10 * time.Millisecond, Logger: slog.New(slog.NewTextHandler(logs, nil))}, func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	waitForLog(t, logs, "deadline exceeded")
}

func TestGoStaysQuietOnShutdown(t *testing.T) {
	logs := &lockedBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	finished := Go(ctx, GoOptions{Logger: slog.New(slog.NewTextHandler(logs, nil))}, func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	cancel()
	<-finished
	if logs.String() != "" {
		t.Fatalf("shutdown was logged as a failure: %q", logs.String())
	}
}
