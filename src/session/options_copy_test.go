package session

import (
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
)

func TestSessionCopiesIgnoredEventsAtConstruction(t *testing.T) {
	ignored := []string{"MESSAGE_CREATE"}
	s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{IgnoredEvents: ignored})
	if err != nil {
		t.Fatal(err)
	}
	ignored[0] = "invalid ignored event"
	options := s.gatewayOptions()
	if len(options.IgnoredEvents) != 1 || options.IgnoredEvents[0] != "MESSAGE_CREATE" {
		t.Errorf("caller changed the stored Gateway configuration: %v", options.IgnoredEvents)
	}
	if err := gateway.ValidateIgnoredEvents(options.IgnoredEvents); err != nil {
		t.Errorf("valid session options became invalid after the constructor returned: %v", err)
	}
}

func TestSessionIgnoredEventsStayStableDuringCallerEdits(t *testing.T) {
	ignored := []string{"MESSAGE_CREATE"}
	s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{IgnoredEvents: ignored})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-start
		for range 2000 {
			ignored[0] = "MESSAGE_DELETE"
		}
	}()
	close(start)
	changed := false
	for range 2000 {
		if s.gatewayOptions().IgnoredEvents[0] != "MESSAGE_CREATE" {
			changed = true
		}
	}
	<-done
	if changed || s.gatewayOptions().IgnoredEvents[0] != "MESSAGE_CREATE" {
		t.Error("caller-owned slice edits changed session event filtering")
	}
}
