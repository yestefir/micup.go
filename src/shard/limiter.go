package shard

import (
	"context"
	"sync"
	"time"
)

type identifyLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func (l *identifyLimiter) wait(ctx context.Context) error {
	l.mu.Lock()
	at := l.next
	if now := time.Now(); at.Before(now) {
		at = now
	}
	l.next = at.Add(l.interval)
	l.mu.Unlock()
	timer := time.NewTimer(time.Until(at))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
