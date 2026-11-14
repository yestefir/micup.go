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

func TestPendingCountRequestsEndWhenSessionRunStops(t *testing.T) {
	for _, channel := range []bool{false, true} {
		name := "guild"
		if channel {
			name = "channel"
		}
		t.Run(name, func(t *testing.T) {
			p := fixture.New(t)
			p.HoldCounts = true
			s := newSession(t, p, session.Options{})
			runCtx, stopRun := context.WithCancel(context.Background())
			runDone := make(chan error, 1)
			go func() { runDone <- s.Run(runCtx); close(runDone) }()
			t.Cleanup(func() { stopRun(); <-runDone })
			fixture.Eventually(t, s.Ready)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				var err error
				if channel {
					_, err = s.RequestChannelMemberCounts(ctx, fixture.GuildID, []string{fixture.VoiceID})
				} else {
					_, err = s.RequestGuildCounts(ctx, []string{fixture.GuildID})
				}
				done <- err
			}()
			fixture.Eventually(t, func() bool { return p.CountsRequested.Load() == 1 })
			stopRun()
			if err := <-runDone; err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, gateway.ErrClosed) {
				t.Errorf("count request outlived Run instead of reporting closed: %v", err)
			}
		})
	}
}
