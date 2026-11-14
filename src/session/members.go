package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

var ErrMembersRateLimited = errors.New("Micup rate limited the member request")

const (
	maxMemberUserIDs          = 100
	maxMemberLimit            = 100
	maxMemberRequestsInFlight = 2
)

type memberChunk struct {
	members     []model.Member
	sequence    int64
	last        bool
	hasSequence bool
	err         error
}

type MembersRateLimitedError struct {
	GuildID    string
	RetryAfter time.Duration
}

func (e *MembersRateLimitedError) Error() string {
	return fmt.Sprintf("Micup rate limited the member request for guild %s; retry after %s", e.GuildID, e.RetryAfter)
}

func (e *MembersRateLimitedError) Is(target error) bool { return target == ErrMembersRateLimited }

func (s *Session) RequestGuildMembers(ctx context.Context, request model.MembersRequest) ([]model.Member, error) {
	if request.GuildID == "" {
		return nil, errors.New("guild ID is required")
	}
	if len(request.UserIDs) > maxMemberUserIDs || request.Limit < 0 || request.Limit > maxMemberLimit {
		return nil, fmt.Errorf("member requests take at most %d user IDs and a limit of 0 to %d", maxMemberUserIDs, maxMemberLimit)
	}
	if !s.Ready() {
		return nil, ErrNotReady
	}
	select {
	case s.memberSlots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.memberSlots }()
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	chunks := make(chan memberChunk, 16)
	s.mu.Lock()
	if !s.Ready() {
		s.mu.Unlock()
		return nil, ErrNotReady
	}
	s.memberRequests[nonce] = chunks
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.memberRequests, nonce)
		s.mu.Unlock()
	}()
	payload := map[string]any{"guild_id": request.GuildID, "query": request.Query, "limit": request.Limit, "presences": request.Presences, "nonce": nonce}
	if len(request.UserIDs) != 0 {
		payload["user_ids"] = request.UserIDs
	}
	if err = s.Send(ctx, gateway.OpRequestGuildMembers, payload); err != nil {
		return nil, err
	}
	var members []model.Member
	var seen map[int64]struct{}
	for {
		select {
		case <-ctx.Done():
			return members, ctx.Err()
		case chunk := <-chunks:
			if chunk.err != nil {
				return members, chunk.err
			}
			if chunk.hasSequence {
				if _, duplicate := seen[chunk.sequence]; duplicate {
					continue
				}
				if seen == nil {
					seen = make(map[int64]struct{})
				}
				seen[chunk.sequence] = struct{}{}
			}
			members = append(members, chunk.members...)
			if chunk.last {
				return members, nil
			}
		}
	}
}

func (s *Session) handleMembersChunk(event gateway.Event) {
	chunk, ok := events.Parse[events.MembersChunk](event.Data)
	if !ok || chunk.Nonce == "" {
		return
	}
	response := memberChunk{members: chunk.Members, last: chunk.ChunkIndex >= chunk.ChunkCount-1}
	if event.Sequence != nil {
		response.sequence, response.hasSequence = *event.Sequence, true
	}
	s.deliverChunk(chunk.Nonce, response)
}

func (s *Session) handleRateLimited(event gateway.Event) {
	limited, ok := events.Parse[events.RateLimited](event.Data)
	if !ok || limited.Opcode != gateway.OpRequestGuildMembers || limited.Meta.Nonce == "" {
		return
	}
	s.deliverChunk(limited.Meta.Nonce, memberChunk{err: &MembersRateLimitedError{GuildID: limited.Meta.GuildID, RetryAfter: time.Duration(limited.RetryAfter * float64(time.Second))}})
}

func (s *Session) deliverChunk(nonce string, chunk memberChunk) {
	s.mu.Lock()
	defer s.mu.Unlock()
	chunks := s.memberRequests[nonce]
	if chunks == nil {
		return
	}
	select {
	case chunks <- chunk:
	default:
		delete(s.memberRequests, nonce)
		select {
		case <-chunks:
		default:
		}
		chunks <- memberChunk{err: errors.New("Micup member response buffer overflow; member list is incomplete")}
	}
}

func (s *Session) stopMemberRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for nonce, chunks := range s.memberRequests {
		delete(s.memberRequests, nonce)
	drain:
		for {
			select {
			case <-chunks:
			default:
				break drain
			}
		}
		chunks <- memberChunk{err: gateway.ErrClosed}
	}
}

func newNonce() (string, error) {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
