package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/session"
)

func TestCountRequestRejectsUnrequestedRows(t *testing.T) {
	wrong, empty := "unrequested", ""
	for _, tc := range []struct {
		name    string
		channel bool
		guild   *string
		id      *string
	}{
		{"guild matching", false, nil, nil},
		{"guild wrong", false, &wrong, nil},
		{"guild empty", false, &empty, nil},
		{"channel matching", true, nil, nil},
		{"channel wrong guild", true, &wrong, nil},
		{"channel empty guild", true, &empty, nil},
		{"channel wrong ID", true, nil, &wrong},
		{"channel empty ID", true, nil, &empty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := fixture.New(t)
			p.CountGuild, p.CountChannel = tc.guild, tc.id
			s := newSession(t, p, session.Options{})
			run(t, s)
			fixture.Eventually(t, s.Ready)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			var rows int
			var err error
			if tc.channel {
				counts, callErr := s.RequestChannelMemberCounts(ctx, fixture.GuildID, []string{fixture.TextID})
				rows, err = len(counts), callErr
			} else {
				counts, callErr := s.RequestGuildCounts(ctx, []string{fixture.GuildID})
				rows, err = len(counts), callErr
			}
			if tc.guild != nil || tc.id != nil {
				if err == nil || rows != 0 {
					t.Fatalf("unrequested row accepted: %d rows, error %v", rows, err)
				}
			} else if err != nil || rows != 1 {
				t.Fatalf("matching response returned %d rows, error %v", rows, err)
			}
		})
	}
}
