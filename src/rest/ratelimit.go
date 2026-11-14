package rest

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	bucketCleanupThreshold = 4096
	bucketCleanupInterval  = time.Second
)

type limiter struct {
	mu          sync.Mutex
	global      time.Time
	buckets     map[string]time.Time
	nextCleanup time.Time
}

func newLimiter() *limiter { return &limiter{buckets: map[string]time.Time{}} }

func (l *limiter) wait(ctx context.Context, route string) error {
	for {
		l.mu.Lock()
		until := l.global
		if bucket := l.buckets[route]; bucket.After(until) {
			until = bucket
		}
		l.mu.Unlock()
		wait := time.Until(until)
		if wait <= 0 {
			return ctx.Err()
		}
		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}

func (l *limiter) observe(route string, header http.Header) {
	if header.Get("X-RateLimit-Remaining") != "0" {
		return
	}
	reset := headerSeconds(header.Get("X-RateLimit-Reset-After"))
	if reset <= 0 {
		return
	}
	l.block(route, false, reset)
}

func (l *limiter) block(route string, global bool, wait time.Duration) {
	now := time.Now()
	until := now.Add(wait)
	l.mu.Lock()
	defer l.mu.Unlock()
	if global {
		if until.After(l.global) {
			l.global = until
		}
		return
	}
	if until.After(l.buckets[route]) {
		l.buckets[route] = until
	}
	// Large maps can stay above the threshold when most buckets are active.
	if len(l.buckets) > bucketCleanupThreshold && !now.Before(l.nextCleanup) {
		now = time.Now()
		l.nextCleanup = now.Add(bucketCleanupInterval)
		for key, bucketUntil := range l.buckets {
			if !bucketUntil.After(now) {
				delete(l.buckets, key)
			}
		}
	}
}

var majorParameters = map[string]bool{"channels": true, "guilds": true, "webhooks": true}

func routeKey(method, path string) string {
	parts := strings.Split(strings.SplitN(path, "?", 2)[0], "/")
	for i := 1; i < len(parts); i++ {
		previous := parts[i-1]
		switch {
		case previous == "reactions":
			parts[i] = "{emoji}"
		case isSnowflake(parts[i]) && !majorParameters[previous]:
			parts[i] = "{id}"
		case previous != "" && i >= 2 && parts[i-2] == "webhooks" && isSnowflake(previous):
			parts[i] = "{token}"
		}
	}
	return method + " " + strings.Join(parts, "/")
}

func isSnowflake(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
