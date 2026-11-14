package session_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

func TestMemberRequestNeedsAllChunkIndexes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		order   []int
		missing bool
		counts  []int
	}{
		{"ordered", []int{0, 1}, false, nil},
		{"final first", []int{1, 0}, false, nil},
		{"missing first", []int{1}, true, nil},
		{"repeated first", []int{0, 0, 1}, false, nil},
		{"repeated final", []int{1, 1, 0}, false, nil},
		{"changed count", []int{0, 1}, false, []int{2, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := fixture.New(t)
			p.MemberChunkOrder = tc.order
			p.MemberChunkCounts = tc.counts
			s := newSession(t, p, session.Options{})
			run(t, s)
			fixture.Eventually(t, s.Ready)
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			members, err := s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID})
			if tc.counts != nil {
				if err == nil || !strings.Contains(err.Error(), "changed its chunk count") || len(members) != 1 {
					t.Fatalf("inconsistent count returned (%d rows, %v), want one partial row and count error", len(members), err)
				}
				return
			}
			if tc.missing {
				if !errors.Is(err, context.DeadlineExceeded) || len(members) != 1 {
					t.Fatalf("missing chunk returned (%d rows, %v), want one partial row and deadline", len(members), err)
				}
				return
			}
			if err != nil || len(members) != 2 {
				t.Fatalf("complete request returned (%d rows, %v), want two rows", len(members), err)
			}
			ids := map[string]bool{}
			for _, member := range members {
				ids[member.User.ID] = true
			}
			if !ids["listener"] || !ids["dj"] {
				t.Fatalf("missing member in completed result: %v", ids)
			}
		})
	}
}
