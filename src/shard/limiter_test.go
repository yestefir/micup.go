package shard

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type identifyWaitContext struct {
	context.Context
	entered chan struct{}
	once    sync.Once
}

func (c *identifyWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.entered) })
	return c.Context.Done()
}

func TestCancelledIdentifyWaitsDoNotDelayLaterShards(t *testing.T) {
	for _, queued := range []bool{false, true} {
		name := "already_cancelled"
		if queued {
			name = "cancelled_while_waiting"
		}
		t.Run(name, func(t *testing.T) {
			// Retain an existing cooldown, but do not wait an hour in the test.
			// Cancelled waiters must not extend when a later shard can identify.
			next := time.Now().Add(time.Hour)
			limiter := newIdentifyLimiter(5 * time.Second)
			limiter.next = next
			for range 4 {
				ctx, cancel := context.WithCancel(context.Background())
				if !queued {
					cancel()
					if err := limiter.wait(ctx); !errors.Is(err, context.Canceled) {
						t.Fatalf("cancelled wait returned %v", err)
					}
					continue
				}
				observed := &identifyWaitContext{Context: ctx, entered: make(chan struct{})}
				done := make(chan error, 1)
				go func() { done <- limiter.wait(observed) }()
				select {
				case <-observed.entered:
				case <-time.After(3 * time.Second):
					cancel()
					t.Fatal("identify call did not reach the wait")
				}
				cancel()
				select {
				case err := <-done:
					if !errors.Is(err, context.Canceled) {
						t.Fatalf("queued wait returned %v", err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("identify wait did not end with its context")
				}
			}
			if extra := limiter.next.Sub(next); extra != 0 {
				t.Fatalf("four cancelled waits added %v to later shards' cooldown", extra)
			}
		})
	}
}

func TestIdentifyWaitCanCancelWhileAdmissionIsBusy(t *testing.T) {
	limiter := newIdentifyLimiter(5 * time.Second)
	limiter.gate <- struct{}{}
	defer func() { <-limiter.gate }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	observed := &identifyWaitContext{Context: ctx, entered: make(chan struct{})}
	done := make(chan error, 1)
	go func() { done <- limiter.wait(observed) }()
	select {
	case <-observed.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("identify call did not reach admission")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("queued admission returned %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("cancelled admission still waits for another caller")
	}
	if !limiter.next.IsZero() {
		t.Fatal("cancelled admission started a cooldown")
	}
}

var benchmarkIdentifyLimiter *identifyLimiter

func BenchmarkIdentifyLimiterConstruction(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkIdentifyLimiter = newIdentifyLimiter(5 * time.Second)
	}
}
