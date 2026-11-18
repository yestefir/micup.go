package session_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

func TestMemberRequestRejectsWrongGuild(t *testing.T) {
	for _, guild := range []string{fixture.GuildID, "different-guild", ""} {
		t.Run("response="+guild, func(t *testing.T) {
			p := fixture.New(t)
			p.MemberGuild = &guild
			s := newSession(t, p, session.Options{})
			run(t, s)
			fixture.Eventually(t, s.Ready)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			members, err := s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID})
			if guild == fixture.GuildID {
				if err != nil || len(members) != 2 {
					t.Fatalf("matching guild returned (%d members, %v)", len(members), err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "guild") || len(members) != 0 {
				t.Fatalf("wrong guild accepted: %d members, error %v", len(members), err)
			}
		})
	}
}
