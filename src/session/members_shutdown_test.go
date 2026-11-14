package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

func TestPendingMemberRequestsEndWhenSessionRunStops(t *testing.T) {
	p := fixture.New(t)
	p.MemberDelay = 500 * time.Millisecond
	s := newSession(t, p, session.Options{})
	runCtx, stopRun := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- s.Run(runCtx)
		close(runDone)
	}()
	t.Cleanup(func() { stopRun(); <-runDone })
	fixture.Eventually(t, s.Ready)
	requestCtx, cancelRequests := context.WithTimeout(context.Background(), time.Second)
	defer cancelRequests()
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := s.RequestGuildMembers(requestCtx, model.MembersRequest{GuildID: fixture.GuildID})
			errs <- err
		}()
	}
	fixture.Eventually(t, func() bool { return p.MemberInFlight.Load() == 2 })
	stopRun()
	if err := <-runDone; err != nil {
		t.Fatalf("session shutdown failed: %v", err)
	}
	for range 2 {
		if err := <-errs; !errors.Is(err, gateway.ErrClosed) {
			t.Errorf("pending member request outlived session instead of reporting closed: %v", err)
		}
	}
	fixture.Eventually(t, func() bool { return p.MemberInFlight.Load() == 0 })
}
