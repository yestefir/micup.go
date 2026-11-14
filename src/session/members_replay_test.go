package session_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

func TestMemberRequestDoesNotAppendReplayedChunks(t *testing.T) {
	for _, mode := range []string{"resume", "edited_event_sequence", "distinct_events_same_user"} {
		t.Run(mode, func(t *testing.T) { testMemberReplay(t, mode) })
	}
}

func testMemberReplay(t *testing.T, mode string) {
	p := fixture.New(t)
	replay := mode != "distinct_events_same_user"
	p.ReplayMemberChunk = replay
	p.RepeatMemberUser = !replay
	var attempts atomic.Int64
	s := newSession(t, p, session.Options{MinBackoff: time.Millisecond, Handler: func(_ context.Context, event gateway.Event) error {
		if event.Name == gateway.EventGuildMembersChunk && attempts.Add(1) == 1 && replay {
			if mode == "edited_event_sequence" {
				*event.Sequence = 999 // Application edits must not change collector identity.
			}
			return errors.New("retry this member chunk through RESUME")
		}
		return nil
	}})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	members, err := s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID})
	if err != nil {
		t.Fatal(err)
	}
	wantAttempts, wantSecond := int64(2), "listener"
	if replay {
		fixture.Eventually(t, func() bool { return p.Resumes.Load() == 1 && s.Ready() })
		wantAttempts, wantSecond = 3, "dj"
	}
	if attempts.Load() != wantAttempts {
		t.Fatalf("raw handler saw %d chunk attempts, want %d", attempts.Load(), wantAttempts)
	}
	if len(members) != 2 || members[0].User.ID != "listener" || members[1].User.ID != wantSecond {
		ids := make([]string, len(members))
		for i, member := range members {
			ids[i] = member.User.ID
		}
		t.Fatalf("member result duplicated a replayed chunk: %v", ids)
	}
}
