package session

import (
	"context"
	"errors"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func resetVoiceSession(t *testing.T, s *Session, ready bool) {
	t.Helper()
	if ready {
		if err := s.dispatch(context.Background(), gateway.Event{Name: gateway.EventReady, Data: []byte(`{"guilds":[]}`)}); err != nil {
			t.Fatal(err)
		}
	} else {
		s.gatewayOptions().OnSessionInvalidated()
	}
}

func TestSessionResetRejectsConsumedVoiceGrant(t *testing.T) {
	for _, ready := range []bool{false, true} {
		name := "invalid_session"
		if ready {
			name = "ready"
		}
		t.Run(name, func(t *testing.T) {
			s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
			if err != nil {
				t.Fatal(err)
			}
			pending, err := s.beginJoin("g")
			if err != nil {
				t.Fatal(err)
			}
			pending <- joinResult{grant: model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", ConnectionID: "old", Endpoint: "wss://voice.invalid", Token: "fixture-grant"}}
			result := <-pending // Caller has received its grant but has not committed it.
			resetVoiceSession(t, s, ready)
			if _, err := s.finishJoin("g", "v", pending, result); !errors.Is(err, ErrVoiceGrant) {
				t.Errorf("old session grant was accepted: %v", err)
			}
			if _, ok := s.Voice("g"); ok {
				t.Error("old grant was restored after the session reset")
			}
		})
	}
}

func TestSessionResetNotifiesEmptyAndBufferedVoiceJoins(t *testing.T) {
	for _, ready := range []bool{false, true} {
		name := "invalid_session"
		if ready {
			name = "ready"
		}
		t.Run(name, func(t *testing.T) {
			s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
			if err != nil {
				t.Fatal(err)
			}
			for _, buffered := range []bool{false, true} {
				pending, err := s.beginJoin("g")
				if err != nil {
					t.Fatal(err)
				}
				if buffered {
					pending <- joinResult{grant: model.VoiceServerUpdate{GuildID: "g"}}
				}
				resetVoiceSession(t, s, ready)
				select {
				case result := <-pending:
					if !errors.Is(result.err, ErrVoiceGrant) {
						t.Errorf("reset did not replace buffered=%v result with invalid grant: %v", buffered, result.err)
					}
				default:
					t.Errorf("reset did not notify buffered=%v pending join", buffered)
				}
				s.endJoin("g", pending)
			}
		})
	}
}

func TestEarlierJoinCleanupKeepsReplacementJoin(t *testing.T) {
	s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.beginJoin("g")
	if err != nil {
		t.Fatal(err)
	}
	resetVoiceSession(t, s, true)
	current, err := s.beginJoin("g")
	if err != nil {
		t.Fatal(err)
	}
	s.endJoin("g", old)
	if s.pending["g"].results != current || len(s.joinOrder) != 1 || s.joinOrder[0] != "g" {
		t.Fatal("earlier caller cleanup removed the replacement join or its refusal ordering")
	}
	grant := model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", ConnectionID: "new", Endpoint: "wss://voice.invalid", Token: "fixture-grant"}
	if _, err := s.finishJoin("g", "v", current, joinResult{grant: grant}); err != nil {
		t.Fatalf("replacement join could not commit its grant: %v", err)
	}
	s.endJoin("g", current)
	if len(s.pending) != 0 || len(s.joinOrder) != 0 {
		t.Error("replacement caller cleanup left a reservation")
	}
}
