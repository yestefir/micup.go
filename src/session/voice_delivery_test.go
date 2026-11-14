package session

import (
	"context"
	"errors"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func TestSessionShutdownNotifiesEmptyAndFullVoiceJoins(t *testing.T) {
	s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	empty, err := s.beginJoin("empty")
	if err != nil {
		t.Fatal(err)
	}
	full, err := s.beginJoin("full")
	if err != nil {
		t.Fatal(err)
	}
	full <- joinResult{grant: model.VoiceServerUpdate{GuildID: "full"}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	for guild, queue := range map[string]chan joinResult{"empty": empty, "full": full} {
		select {
		case result := <-queue:
			if !errors.Is(result.err, gateway.ErrClosed) {
				t.Errorf("%s join did not receive the closed error: %+v", guild, result)
			}
		default:
			t.Errorf("%s join received no shutdown notification", guild)
		}
		if _, err := s.beginJoin(guild); !errors.Is(err, ErrVoiceJoinPending) {
			t.Errorf("guild reopened before the previous caller finished cleanup: %v", err)
		}
		s.endJoin(guild, queue)
	}
	if len(s.pending) != 0 || len(s.joinOrder) != 0 {
		t.Error("join state remained after callers finished cleanup")
	}
}
