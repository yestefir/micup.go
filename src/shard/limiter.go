package shard

import (
	"context"
	"time"
)

type identifyLimiter struct {
	gate     chan struct{}
	interval time.Duration
	next     time.Time
}

func newIdentifyLimiter(interval time.Duration) *identifyLimiter {
	return &identifyLimiter{gate: make(chan struct{}, 1), interval: interval}
}

func (l *identifyLimiter) wait(ctx context.Context) error {
	// Hold admission through the cooldown so cancelled waiters do not reserve
	// future slots. Waiting for admission itself must also be cancellable.
	select {
	case l.gate <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-l.gate }()
	if wait := time.Until(l.next); wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	l.next = time.Now().Add(l.interval)
	return nil
}
