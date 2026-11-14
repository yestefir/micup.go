package session_test

import (
	"context"
	"sync"
	"testing"

	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

func TestSessionCopiesInitialCustomStatus(t *testing.T) {
	initial := &model.Presence{Status: model.StatusIdle, CustomStatus: &model.CustomStatus{Text: "initial"}}
	s, err := session.New(session.Config{Endpoint: "wss://localhost", Token: "fixture"}, session.Options{Presence: initial})
	if err != nil {
		t.Fatal(err)
	}
	initial.CustomStatus.Text = "caller changed"
	if got := s.Presence().CustomStatus.Text; got != "initial" {
		t.Fatalf("caller changed session custom status through options: %q", got)
	}
}

func TestSessionPresenceReturnsIndependentCustomStatus(t *testing.T) {
	s, err := session.New(session.Config{Endpoint: "wss://localhost", Token: "fixture"}, session.Options{Presence: &model.Presence{Status: model.StatusIdle, CustomStatus: &model.CustomStatus{Text: "initial"}}})
	if err != nil {
		t.Fatal(err)
	}
	snapshot := s.Presence()
	snapshot.CustomStatus.Text = "snapshot changed"
	if got := s.Presence().CustomStatus.Text; got != "initial" {
		t.Errorf("snapshot changed session custom status: %q", got)
	}
	var writers sync.WaitGroup
	for range 8 {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for range 100 {
				own := s.Presence()
				own.CustomStatus.Text = "own snapshot"
			}
		}()
	}
	writers.Wait()
	if got := s.Presence().CustomStatus.Text; got != "initial" {
		t.Fatalf("concurrent snapshots changed stored custom status: %q", got)
	}
}

func TestSessionCopiesUpdatedCustomStatus(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	custom := &model.CustomStatus{Text: "updated"}
	if err := s.UpdatePresence(context.Background(), model.Presence{Status: model.StatusDND, CustomStatus: custom}); err != nil {
		t.Fatal(err)
	}
	custom.Text = "caller changed"
	if got := s.Presence().CustomStatus.Text; got != "updated" {
		t.Fatalf("caller changed stored presence after UpdatePresence returned: %q", got)
	}
}

var benchmarkPresence model.Presence

func BenchmarkSessionPresenceSnapshot(b *testing.B) {
	for _, custom := range []bool{false, true} {
		name := "plain"
		presence := model.Presence{Status: model.StatusOnline}
		if custom {
			name = "custom"
			presence.CustomStatus = &model.CustomStatus{Text: "playing"}
		}
		b.Run(name, func(b *testing.B) {
			s, err := session.New(session.Config{Endpoint: "wss://localhost", Token: "fixture"}, session.Options{Presence: &presence})
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				benchmarkPresence = s.Presence()
			}
		})
	}
}
