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

func TestMemberRateLimitRejectsWrongGuild(t *testing.T) {
	for _, guild := range []string{fixture.GuildID, "different-guild", ""} {
		t.Run("response="+guild, func(t *testing.T) {
			p := fixture.New(t)
			p.RateLimitMembers = true
			p.MemberGuild = &guild
			s := newSession(t, p, session.Options{})
			run(t, s)
			fixture.Eventually(t, s.Ready)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			members, err := s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID})
			if len(members) != 0 {
				t.Fatalf("rate-limited request returned %d rows", len(members))
			}
			if guild == fixture.GuildID || guild == "" {
				var limited *session.MembersRateLimitedError
				if !errors.As(err, &limited) || limited.RetryAfter != 12500*time.Millisecond {
					t.Fatalf("matching response lost rate-limit advice: %v", err)
				}
			} else if err == nil || errors.Is(err, session.ErrMembersRateLimited) || !strings.Contains(err.Error(), "mismatched guild") {
				t.Fatalf("wrong-guild rate-limit advice accepted: %v", err)
			}
		})
	}
}
