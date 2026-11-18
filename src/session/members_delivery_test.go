package session

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func TestMemberChunkOverflowFailsInsteadOfDroppingData(t *testing.T) {
	for _, terminal := range []bool{false, true} {
		t.Run(fmt.Sprint(terminal), func(t *testing.T) {
			chunks := make(chan memberChunk, 16)
			other := make(chan memberChunk, 1)
			s := &Session{memberRequests: map[string]chan memberChunk{"request": chunks, "other": other}}
			for i := range cap(chunks) + 1 {
				s.deliverChunk("request", memberChunk{
					members: []model.Member{{User: &model.User{ID: fmt.Sprint(i)}}},
					last:    terminal && i == cap(chunks),
				})
			}
			var failure error
			for len(chunks) > 0 {
				chunk := <-chunks
				if chunk.err != nil {
					failure = chunk.err
				}
				if chunk.last && chunk.err == nil {
					t.Error("overflow must not report a successful complete result")
				}
			}
			if failure == nil {
				t.Error("overflow silently lost response data without failing the request")
			}
			s.deliverChunk("request", memberChunk{last: true})
			select {
			case chunk := <-chunks:
				t.Errorf("failed request accepted a later response: %+v", chunk)
			default:
			}
			s.deliverChunk("other", memberChunk{last: true})
			if chunk := <-other; chunk.err != nil || !chunk.last {
				t.Fatalf("overflow affected another request: %+v", chunk)
			}
		})
	}
}

func TestSessionShutdownFailsEmptyAndFullMemberQueues(t *testing.T) {
	empty := make(chan memberChunk, 16)
	full := make(chan memberChunk, 16)
	for i := range cap(full) {
		full <- memberChunk{last: i == cap(full)-1}
	}
	s := &Session{memberRequests: map[string]chan memberChunk{"empty": empty, "full": full}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	for name, queue := range map[string]chan memberChunk{"empty": empty, "full": full} {
		select {
		case chunk := <-queue:
			if !errors.Is(chunk.err, gateway.ErrClosed) || chunk.last {
				t.Errorf("%s request did not receive shutdown error: %+v", name, chunk)
			}
		default:
			t.Errorf("%s request did not receive a shutdown notification", name)
		}
	}
	if len(s.memberRequests) != 0 {
		t.Error("shutdown retained pending request registrations")
	}
}

func TestMemberChunksWithinCapacityRemainOrdered(t *testing.T) {
	chunks := make(chan memberChunk, 16)
	other := make(chan memberChunk, 16)
	s := &Session{memberRequests: map[string]chan memberChunk{"request": chunks, "other": other}}
	for i := range cap(chunks) {
		s.deliverChunk("request", memberChunk{members: []model.Member{{User: &model.User{ID: fmt.Sprint(i)}}}, last: i == cap(chunks)-1})
	}
	s.deliverChunk("other", memberChunk{last: true})
	for i := range cap(chunks) {
		chunk := <-chunks
		if chunk.err != nil || len(chunk.members) != 1 || chunk.members[0].User.ID != fmt.Sprint(i) || chunk.last != (i == cap(chunks)-1) {
			t.Fatalf("normal response %d changed: %+v", i, chunk)
		}
	}
	if chunk := <-other; chunk.err != nil || !chunk.last {
		t.Fatalf("another request was affected: %+v", chunk)
	}
	s.deliverChunk("cancelled", memberChunk{last: true})
}
