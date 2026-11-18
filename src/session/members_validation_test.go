package session

import (
	"fmt"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
)

func TestMemberChunkRejectsInvalidPosition(t *testing.T) {
	for _, tc := range []struct {
		name  string
		index int
		count int
	}{
		{"zero count", 0, 0},
		{"negative count", 0, -1},
		{"negative index", -1, 2},
		{"index equals count", 2, 2},
		{"index exceeds count", 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chunks := make(chan memberChunk, 1)
			s := &Session{memberRequests: map[string]chan memberChunk{"request": chunks}}
			s.handleMembersChunk(gateway.Event{Data: []byte(fmt.Sprintf(
				`{"guild_id":"guild","nonce":"request","chunk_index":%d,"chunk_count":%d,"members":[{"user":{"id":"member"}}]}`,
				tc.index, tc.count,
			))})
			select {
			case chunk := <-chunks:
				if chunk.err == nil || chunk.last || len(chunk.members) != 0 {
					t.Errorf("invalid chunk accepted: %+v", chunk)
				}
			default:
				t.Error("invalid chunk left the request waiting without an error")
			}
		})
	}
}

func TestMemberChunkAcceptsValidPosition(t *testing.T) {
	for _, index := range []int{0, 1} {
		chunks := make(chan memberChunk, 1)
		s := &Session{memberRequests: map[string]chan memberChunk{"request": chunks}}
		sequence := int64(index + 10)
		s.handleMembersChunk(gateway.Event{Sequence: &sequence, Data: []byte(fmt.Sprintf(
			`{"guild_id":"guild","nonce":"request","chunk_index":%d,"chunk_count":2,"members":[{"user":{"id":"member"}}]}`, index,
		))})
		select {
		case chunk := <-chunks:
			if chunk.err != nil || chunk.last != (index == 1) || len(chunk.members) != 1 || !chunk.hasSequence || chunk.sequence != sequence {
				t.Errorf("valid chunk changed: %+v", chunk)
			}
		default:
			t.Error("valid chunk was ignored")
		}
	}
}
