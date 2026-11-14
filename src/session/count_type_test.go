package session_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/session"
)

func TestCountRequestChecksResponseType(t *testing.T) {
	for _, method := range []string{"guild", "channel"} {
		for _, swapped := range []bool{false, true} {
			name := method + "/matching"
			if swapped {
				name = method + "/swapped"
			}
			t.Run(name, func(t *testing.T) {
				p := fixture.New(t)
				p.SwapCountEvents = swapped
				s := newSession(t, p, session.Options{})
				run(t, s)
				fixture.Eventually(t, s.Ready)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var rows int
				if method == "guild" {
					counts, callErr := s.RequestGuildCounts(ctx, []string{fixture.GuildID})
					rows, err = len(counts), callErr
				} else {
					counts, callErr := s.RequestChannelMemberCounts(ctx, fixture.GuildID, []string{fixture.TextID})
					rows, err = len(counts), callErr
				}
				if swapped {
					if err == nil || !strings.Contains(err.Error(), "carries") || rows != 0 {
						t.Fatalf("wrong event type returned %d rows, error %v", rows, err)
					}
				} else if err != nil || rows != 1 {
					t.Fatalf("matching event returned %d rows, error %v", rows, err)
				}
			})
		}
	}
}
