package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/session"
)

func TestPendingVoiceJoinRespectsSessionReplacement(t *testing.T) {
	for _, mode := range []string{"invalid_session", "new_ready", "resume"} {
		t.Run(mode, func(t *testing.T) {
			p := fixture.New(t)
			p.HoldVoice = true
			s := newSession(t, p, session.Options{MinBackoff: time.Millisecond})
			run(t, s)
			fixture.Eventually(t, s.Ready)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{})
				done <- err
			}()
			fixture.Eventually(t, func() bool { return p.VoiceJoins.Load() == 1 })
			switch mode {
			case "invalid_session":
				p.Send(map[string]any{"op": gateway.OpInvalidSession, "d": false})
			case "new_ready":
				p.Event(gateway.EventReady, map[string]any{"session_id": "replacement", "guilds": []any{}})
			case "resume":
				p.Send(map[string]any{"op": gateway.OpReconnect})
				fixture.Eventually(t, func() bool { return p.Resumes.Load() == 1 && s.Ready() })
				p.Event(gateway.EventVoiceServerUpdate, map[string]string{"guild_id": fixture.GuildID, "channel_id": fixture.VoiceID, "connection_id": fixture.Connection, "endpoint": "wss://voice.invalid", "token": "fixture-grant"})
			}
			err := <-done
			if mode == "resume" {
				if err != nil {
					t.Fatalf("resumable reconnect invalidated the join: %v", err)
				}
			} else {
				if !errors.Is(err, session.ErrVoiceGrant) {
					t.Errorf("replacement left the old join waiting instead of invalidating its grant: %v", err)
				}
				if _, ok := s.Voice(fixture.GuildID); ok {
					t.Error("replacement retained an old voice grant")
				}
			}
		})
	}
}
