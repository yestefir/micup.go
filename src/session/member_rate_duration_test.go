package session

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
)

func TestMemberRateLimitDurationBounds(t *testing.T) {
	const maxDuration = time.Duration(1<<63 - 1)
	for _, tc := range []struct {
		seconds string
		want    time.Duration
	}{
		{"12.5", 12500 * time.Millisecond},
		{"0", 0},
		{"-1", 0},
		{"-1e100", 0},
		{"1e-12", 0},
		{"1e100", maxDuration},
		{"9223372036.854776", maxDuration},
	} {
		t.Run(tc.seconds, func(t *testing.T) {
			chunks := make(chan memberChunk, 1)
			s := &Session{memberRequests: map[string]chan memberChunk{"request": chunks}}
			s.handleRateLimited(gateway.Event{Data: []byte(fmt.Sprintf(
				`{"opcode":%d,"retry_after":%s,"meta":{"guild_id":"guild","nonce":"request"}}`,
				gateway.OpRequestGuildMembers, tc.seconds,
			))})
			select {
			case chunk := <-chunks:
				var limited *MembersRateLimitedError
				if !errors.As(chunk.err, &limited) || !errors.Is(chunk.err, ErrMembersRateLimited) {
					t.Fatalf("expected member rate-limit error, got %v", chunk.err)
				}
				if limited.RetryAfter != tc.want || limited.GuildID != "guild" {
					t.Errorf("rate-limit error = %+v, want retry delay %s", limited, tc.want)
				}
			default:
				t.Fatal("rate-limit response was not delivered")
			}
		})
	}
}
