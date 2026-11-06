package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestAPIErrorCarriesCodeFieldsAndRequestID(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "req-1")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"code":"INVALID_FORM_BODY","message":"Invalid form body.","errors":[{"path":"embeds.0.title","code":"STRING_LENGTH_INVALID","message":"too long"}]}`)
	})
	err := c.Do(context.Background(), http.MethodPost, "/channels/1/messages", map[string]string{"content": "x"}, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "INVALID_FORM_BODY" || apiErr.RequestID != "req-1" || len(apiErr.Errors) != 1 || apiErr.Errors[0].Path != "embeds.0.title" {
		t.Fatalf("error details were lost: %#v", err)
	}
	if apiErr.Error() != "Micup API returned HTTP 400 (INVALID_FORM_BODY): Invalid form body.; embeds.0.title: too long request req-1" {
		t.Fatalf("unexpected error text: %s", apiErr.Error())
	}
}

func TestGlobalRateLimitPausesEveryRoute(t *testing.T) {
	var calls atomic.Int64
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"code":"RATE_LIMITED","global":true,"retry_after":0.2}`)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	done := make(chan time.Time, 1)
	go func() {
		_ = c.Do(context.Background(), http.MethodGet, "/channels/1", nil, nil)
		done <- time.Now()
	}()
	time.Sleep(50 * time.Millisecond)
	started := time.Now()
	if err := c.Do(context.Background(), http.MethodGet, "/guilds/2", nil, nil); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) < 100*time.Millisecond {
		t.Fatal("a different route ignored the global rate limit")
	}
	<-done
}

func TestBucketHeadersDelayTheNextRequest(t *testing.T) {
	var calls atomic.Int64
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset-After", "0.2")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Do(context.Background(), http.MethodGet, "/channels/1/messages/5", nil, nil); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if err := c.Do(context.Background(), http.MethodGet, "/channels/1/messages/6", nil, nil); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) < 150*time.Millisecond {
		t.Fatal("exhausted bucket was not respected")
	}
	started = time.Now()
	if err := c.Do(context.Background(), http.MethodGet, "/channels/2/messages/6", nil, nil); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) > 100*time.Millisecond {
		t.Fatal("another channel's bucket was delayed")
	}
}

func TestServerErrorsRetryOnlySafeMethods(t *testing.T) {
	var calls atomic.Int64
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0.01")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Do(context.Background(), http.MethodGet, "/channels/1", nil, nil); err != nil || calls.Load() != 2 {
		t.Fatalf("GET was not retried: %v %d", err, calls.Load())
	}
	calls.Store(0)
	var apiErr *APIError
	if err := c.Do(context.Background(), http.MethodPost, "/channels/1/messages", map[string]string{"content": "x"}, nil); !errors.As(err, &apiErr) || apiErr.StatusCode != 503 || calls.Load() != 1 {
		t.Fatalf("POST without a nonce was retried: %v %d", err, calls.Load())
	}
}

func TestLockedResourceIsRetried(t *testing.T) {
	var calls atomic.Int64
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0.01")
			w.WriteHeader(http.StatusLocked)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	start := time.Now()
	if err := c.Do(context.Background(), http.MethodPatch, "/guilds/1", map[string]string{}, nil); err != nil || calls.Load() != 2 {
		t.Fatalf("locked resource was not retried: %v %d", err, calls.Load())
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Fatal("locked retry did not wait at least one second")
	}
}

func TestAuditReasonHeader(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Audit-Log-Reason") != "spam" {
			t.Errorf("reason header missing: %q", r.Header.Get("X-Audit-Log-Reason"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.call(context.Background(), &request{method: http.MethodDelete, path: "/channels/1/messages/2", reason: "spam"}); err != nil {
		t.Fatal(err)
	}
}

func TestRouteKeyKeepsMajorParameters(t *testing.T) {
	for in, want := range map[string]string{
		"/channels/1/messages/2":                   "GET /channels/1/messages/{id}",
		"/channels/1/messages/2/reactions/%F0/@me": "GET /channels/1/messages/{id}/reactions/{emoji}/@me",
		"/guilds/9/members/3":                      "GET /guilds/9/members/{id}",
		"/webhooks/4/secret":                       "GET /webhooks/4/{token}",
	} {
		if got := routeKey(http.MethodGet, in); got != want {
			t.Fatalf("routeKey(%s) = %s, want %s", in, got, want)
		}
	}
}
