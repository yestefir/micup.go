package rest

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/model"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func loggingClient(t *testing.T, code string, logger *slog.Logger) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, `{"code":%q,"message":"denied"}`, code)
	}))
	t.Cleanup(server.Close)
	c, err := New(Config{BaseURL: server.URL + "/api", Token: "fixture", Logger: logger})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMissingPermissionsAreLoggedOncePerRoute(t *testing.T) {
	out := &lockedBuffer{}
	c := loggingClient(t, CodeMissingPermissions, slog.New(slog.NewTextHandler(out, nil)))
	ctx := context.Background()
	for range 3 {
		if err := c.BanMember(ctx, "g1", "u1", model.Ban{}); err == nil {
			t.Fatal("expected an error")
		}
	}
	log := out.String()
	if strings.Count(log, "bot lacks permission") != 1 {
		t.Fatalf("log lines = %q, want exactly one", log)
	}
	for _, want := range []string{"level=WARN", "guild=g1", "code=MISSING_PERMISSIONS", "method=PUT"} {
		if !strings.Contains(log, want) {
			t.Errorf("log %q is missing %q", log, want)
		}
	}
	if err := c.DeleteChannel(ctx, "c1"); err == nil {
		t.Fatal("expected an error")
	}
	if strings.Count(out.String(), "bot lacks permission") != 2 {
		t.Fatalf("a different route was not logged: %q", out.String())
	}
}

func TestOtherForbiddenCodesAndNilLoggerStaySilent(t *testing.T) {
	out := &lockedBuffer{}
	c := loggingClient(t, CodeUnknownChannel, slog.New(slog.NewTextHandler(out, nil)))
	if err := c.DeleteChannel(context.Background(), "c1"); err == nil || out.String() != "" {
		t.Fatalf("unrelated 403 logged %q (err %v)", out.String(), err)
	}
	silent := loggingClient(t, CodeMissingPermissions, nil)
	if err := silent.DeleteChannel(context.Background(), "c1"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestPermissionLogWindowExpires(t *testing.T) {
	var p permissionLog
	now := time.Now()
	if !p.allow("k", now) || p.allow("k", now.Add(time.Second)) || !p.allow("k", now.Add(2*permissionLogWindow)) {
		t.Fatal("window did not suppress and then release repeated keys")
	}
}
