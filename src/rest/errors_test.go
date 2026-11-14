package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHeaderSecondsBounds(t *testing.T) {
	for _, item := range []struct {
		value string
		want  time.Duration
	}{
		{"", 0},
		{"invalid", 0},
		{"NaN", 0},
		{"-Inf", 0},
		{"-1e309", 0},
		{"-1", 0},
		{"0", 0},
		{"0.0000000001", 0},
		{"0.25", 250 * time.Millisecond},
		{"30.5", 30500 * time.Millisecond},
		{"9223372036", 9223372036 * time.Second},
		{"9223372037", time.Duration(1<<63 - 1)},
		{"1e100", time.Duration(1<<63 - 1)},
		{"1e309", time.Duration(1<<63 - 1)},
		{"+Inf", time.Duration(1<<63 - 1)},
	} {
		t.Run(item.value, func(t *testing.T) {
			if got := headerSeconds(item.value); got != item.want {
				t.Errorf("headerSeconds(%q) = %s, want %s", item.value, got, item.want)
			}
		})
	}
}

func TestAPIErrorRetryAfterBounds(t *testing.T) {
	for _, item := range []struct {
		body   string
		header string
		want   time.Duration
	}{
		{`{"retry_after":0.25}`, "", 250 * time.Millisecond},
		{`{"retry_after":0.0000000001}`, "", minRetryAfter},
		{`{"retry_after":1e100}`, "", time.Duration(1<<63 - 1)},
		{`{"retry_after":-1}`, "0.25", 250 * time.Millisecond},
		{`{}`, "1e100", time.Duration(1<<63 - 1)},
	} {
		t.Run(item.body+item.header, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{item.header}},
				Body:       io.NopCloser(strings.NewReader(item.body)),
			}
			if got := readAPIError(resp).RetryAfter; got != item.want {
				t.Errorf("RetryAfter = %s, want %s", got, item.want)
			}
		})
	}
}

func TestOversizedRetryAfterDoesNotRetry(t *testing.T) {
	for _, source := range []string{"body", "header"} {
		t.Run(source, func(t *testing.T) {
			var calls atomic.Int64
			c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) > 1 {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if source == "header" {
					w.Header().Set("Retry-After", "1e100")
				}
				w.WriteHeader(http.StatusTooManyRequests)
				if source == "body" {
					fmt.Fprint(w, `{"retry_after":1e100}`)
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := c.Do(ctx, http.MethodGet, "/channels/1", nil, nil)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests || apiErr.RetryAfter <= maxRetryAfter {
				t.Errorf("oversized retry delay was not returned as an API error: %v", err)
			}
			if got := calls.Load(); got != 1 {
				t.Errorf("made %d requests despite oversized Retry-After, want 1", got)
			}
		})
	}
}

func TestOversizedBucketResetHonorsCancellation(t *testing.T) {
	var calls atomic.Int64
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset-After", "1e100")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Do(context.Background(), http.MethodGet, "/channels/1", nil, nil); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := c.Do(ctx, http.MethodGet, "/channels/1", nil, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("bucket wait ignored cancellation: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("sent %d requests despite exhausted bucket, want 1", got)
	}
}

func BenchmarkHeaderSeconds(b *testing.B) {
	for _, value := range []string{"0.25", "1e100"} {
		b.Run(value, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				headerSeconds(value)
			}
		})
	}
}
