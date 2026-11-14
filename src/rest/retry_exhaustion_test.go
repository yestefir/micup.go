package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type retryTransport func(*http.Request) (*http.Response, error)

func (f retryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type retryCloseBody struct {
	io.ReadCloser
	onClose func()
}

func (b retryCloseBody) Close() error {
	err := b.ReadCloser.Close()
	b.onClose()
	return err
}

func TestRetryExhaustionPreservesFinalResponseAndGlobalLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	c, err := New(Config{BaseURL: "https://fixture.invalid/api", Token: "fixture", HTTP: &http.Client{Transport: retryTransport(func(*http.Request) (*http.Response, error) {
		calls++
		body := `{"code":"RATE_LIMITED","retry_after":0.001}`
		if calls == maxAttempts {
			body = `{"code":"RATE_LIMITED","retry_after":1,"global":true}`
		}
		reader := retryCloseBody{ReadCloser: io.NopCloser(strings.NewReader(body)), onClose: func() {
			if calls == maxAttempts {
				cancel()
			}
		}}
		return &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}, Body: reader}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Do(ctx, http.MethodGet, "/test", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests || !apiErr.Global {
		t.Errorf("final API response was hidden by an exhausted retry wait: %v", err)
	}
	if calls != maxAttempts {
		t.Errorf("made %d requests, want %d", calls, maxAttempts)
	}
	c.limits.mu.Lock()
	blocked := !c.limits.global.IsZero()
	c.limits.mu.Unlock()
	if !blocked {
		t.Error("final rate-limit response did not update global pacing")
	}
}

func BenchmarkRESTExhaustedRetries(b *testing.B) {
	for _, mode := range []string{"server", "transport"} {
		b.Run(mode, func(b *testing.B) {
			calls := 0
			c, err := New(Config{BaseURL: "https://fixture.invalid/api", Token: "fixture", HTTP: &http.Client{Transport: retryTransport(func(*http.Request) (*http.Response, error) {
				calls++
				if mode == "transport" {
					return nil, errors.New("fixture connection failure")
				}
				return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"code":"SERVER_FAILURE"}`))}, nil
			})}})
			if err != nil {
				b.Fatal(err)
			}
			for b.Loop() {
				calls = 0
				err = c.Do(context.Background(), http.MethodGet, "/test", nil, nil)
				if err == nil || calls != maxAttempts {
					b.Fatalf("retry workload changed: error %v; calls %d", err, calls)
				}
				if mode == "server" {
					var apiErr *APIError
					if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
						b.Fatalf("last server error was lost: %v", err)
					}
				} else if err.Error() != "Micup API connection failed" {
					b.Fatalf("transport failure changed: %v", err)
				}
			}
			b.ReportMetric(float64(calls), "requests/op")
		})
	}
}
