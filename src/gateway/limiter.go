package gateway

import (
	"context"
	"sync"
	"time"
)

const (
	commandBudget  = 540
	commandWindow  = 60 * time.Second
	presenceBudget = 5
	presenceWindow = 20 * time.Second
	maxPayloadSize = 16384
)

type windowLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	sent   []time.Time
}

func newWindowLimiter(limit int, window time.Duration) *windowLimiter {
	return &windowLimiter{limit: limit, window: window}
}

func (l *windowLimiter) wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-l.window)
		kept := l.sent[:0]
		for _, at := range l.sent {
			if at.After(cutoff) {
				kept = append(kept, at)
			}
		}
		l.sent = kept
		if len(l.sent) < l.limit {
			l.sent = append(l.sent, now)
			l.mu.Unlock()
			return nil
		}
		wait := l.sent[0].Add(l.window).Sub(now)
		l.mu.Unlock()
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
