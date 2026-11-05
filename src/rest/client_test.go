package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c, err := New(Config{BaseURL: server.URL + "/api", Token: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewRejectsInvalidToken(t *testing.T) {
	for _, token := range []string{"", "a\r\nb"} {
		if _, err := New(Config{BaseURL: "https://micup.gg/api", Token: token}); err == nil {
			t.Fatalf("token accepted: %q", token)
		}
	}
}

func TestDoSendsBotAuthorization(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/@me" || r.Header.Get("Authorization") != "Bot fixture" {
			t.Errorf("unexpected request %s %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		fmt.Fprint(w, `{"id":"bot-id","bot":true}`)
	})
	user, err := c.CurrentUser(context.Background())
	if err != nil || user.ID != "bot-id" || !user.Bot {
		t.Fatalf("user not decoded: %+v %v", user, err)
	}
}

func TestDoRejectsNonRelativePath(t *testing.T) {
	c := &Client{baseURL: "https://micup.example/api/v1"}
	for _, p := range []string{"https://attacker.test/", "//attacker.test/"} {
		if err := c.Do(context.Background(), http.MethodGet, p, nil, nil); err == nil {
			t.Fatalf("invalid path accepted: %s", p)
		}
	}
}

func TestDoDoesNotFollowRedirectsWithCredentials(t *testing.T) {
	var leaked atomic.Bool
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Store(r.Header.Get("Authorization") != "") }))
	defer other.Close()
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, other.URL, http.StatusFound) })
	var apiErr *APIError
	if err := c.Do(context.Background(), http.MethodGet, "/users/@me", nil, nil); !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusFound {
		t.Fatalf("redirect was not reported: %v", err)
	}
	if leaked.Load() {
		t.Fatal("credential leaked through redirect")
	}
}

func TestRateLimitHonorsCancellation(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"retry_after":3}`)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := c.Do(ctx, http.MethodGet, "/test", nil, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("rate limit was not cancelled: %v", err)
	}
}

func TestMessageContentGrantFlags(t *testing.T) {
	flags := func(v uint64) *uint64 { return &v }
	for name, item := range map[string]struct {
		app            Application
		enabled, known bool
	}{
		"other-intents": {Application{Flags: flags(40960)}, false, true},
		"limited":       {Application{Flags: flags(524288)}, true, true},
		"approved":      {Application{Flags: flags(262144)}, true, true},
		"unavailable":   {Application{}, false, false},
	} {
		got := item.app.MessageContentEnabled()
		if (got != nil) != item.known || (item.known && *got != item.enabled) {
			t.Fatalf("%s: incorrect Message Content grant", name)
		}
	}
}

func TestMessageBodyValidation(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid message was sent") })
	if _, err := c.SendMessage(context.Background(), "channel", "", ""); err == nil {
		t.Fatal("empty message accepted")
	}
	if _, err := c.SendMessage(context.Background(), "", "hi", ""); err == nil {
		t.Fatal("missing channel accepted")
	}
}
