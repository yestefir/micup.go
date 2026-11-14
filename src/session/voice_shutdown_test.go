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

func TestPendingVoiceJoinEndsWhenSessionRunStops(t *testing.T) {
	p := fixture.New(t)
	p.HoldVoice = true
	s := newSession(t, p, session.Options{})
	runCtx, stopRun := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- s.Run(runCtx)
		close(runDone)
	}()
	t.Cleanup(func() { stopRun(); <-runDone })
	fixture.Eventually(t, s.Ready)
	joinCtx, cancelJoin := context.WithTimeout(context.Background(), time.Second)
	defer cancelJoin()
	joinDone := make(chan error, 1)
	go func() {
		_, err := s.JoinVoice(joinCtx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{})
		joinDone <- err
	}()
	fixture.Eventually(t, func() bool { return p.VoiceJoins.Load() == 1 })
	stopRun()
	if err := <-runDone; err != nil {
		t.Fatalf("session shutdown failed: %v", err)
	}
	if err := <-joinDone; !errors.Is(err, gateway.ErrClosed) {
		t.Errorf("voice join outlived session instead of reporting closed: %v", err)
	}
}
