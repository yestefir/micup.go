package rest

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"testing/synctest"
	"time"
)

func TestLimiterThrottlesExpiredBucketCleanup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := newLimiter()
		for i := range 4096 {
			l.buckets[strconv.Itoa(i)] = time.Now().Add(time.Hour)
		}
		l.block("short", false, time.Millisecond)
		time.Sleep(2 * time.Millisecond)
		l.block("other", false, time.Hour)
		if _, ok := l.buckets["short"]; !ok {
			t.Error("a second full sweep ran inside the cleanup interval")
		}
		before := time.Now()
		if err := l.wait(context.Background(), "short"); err != nil || time.Now() != before {
			t.Errorf("retained expired bucket delayed admission: %v", err)
		}
		time.Sleep(time.Second)
		l.block("other", false, time.Hour)
		if _, ok := l.buckets["short"]; ok {
			t.Error("expired bucket survived the next cleanup")
		}
		if got := len(l.buckets); got != 4097 {
			t.Errorf("got %d buckets after cleanup, want 4097 active buckets", got)
		}
	})
}

func BenchmarkLimiterBlockActiveBuckets(b *testing.B) {
	for _, count := range []int{64, 8192} {
		b.Run(fmt.Sprintf("%d-routes", count), func(b *testing.B) {
			l := newLimiter()
			routes := make([]string, count)
			for i := range routes {
				routes[i] = strconv.Itoa(i)
				l.buckets[routes[i]] = time.Now().Add(time.Hour)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				l.block(routes[i%count], false, time.Hour)
			}
		})
	}
}
