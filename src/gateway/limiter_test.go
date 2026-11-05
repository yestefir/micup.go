package gateway

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestWindowLimiterWaitsForTheOldestSend(t *testing.T) {
	l := newWindowLimiter(2, 100*time.Millisecond)
	ctx := context.Background()
	start := time.Now()
	for range 3 {
		if err := l.wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if time.Since(start) < 90*time.Millisecond {
		t.Fatal("third send did not wait for the window")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_ = l.wait(ctx)
	if err := l.wait(cancelled); err == nil {
		t.Fatal("cancelled wait returned no error")
	}
}

func TestConnRejectsOversizedPayload(t *testing.T) {
	c := newConn(nil)
	err := c.Send(context.Background(), OpPresenceUpdate, map[string]string{"text": strings.Repeat("x", maxPayloadSize)})
	if err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("oversized payload was not rejected: %v", err)
	}
}

func TestIgnoredEventsValidation(t *testing.T) {
	if ValidateIgnoredEvents([]string{"TYPING_START", "MESSAGE_REACTION_ADD"}) != nil {
		t.Fatal("valid events rejected")
	}
	for _, bad := range [][]string{{""}, {"typing_start"}, {"A B"}, make([]string, 257)} {
		if ValidateIgnoredEvents(bad) == nil {
			t.Fatalf("invalid ignored events accepted: %v", bad[:1])
		}
	}
}
