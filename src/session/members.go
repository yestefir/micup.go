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
	guildID     string
	sequence    int64
	index       int
	count       int
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
	indexes := make(map[int]struct{})
	var count int
	for {
		select {
		case <-ctx.Done():
			return members, ctx.Err()
		case chunk := <-chunks:
			if chunk.err != nil {
				var limited *MembersRateLimitedError
				if errors.As(chunk.err, &limited) && limited.GuildID != "" && limited.GuildID != request.GuildID {
					return members, errors.New("Micup member rate-limit response has a mismatched guild")
				}
				return members, chunk.err
			}
			if chunk.guildID != request.GuildID {
				return members, errors.New("Micup member response has a mismatched guild; member list is incomplete")
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
			if count == 0 {
				count = chunk.count
			}
			if chunk.count != count {
				return members, errors.New("Micup member response changed its chunk count; member list is incomplete")
			}
			if _, duplicate := indexes[chunk.index]; duplicate {
				continue
			}
			indexes[chunk.index] = struct{}{}
			members = append(members, chunk.members...)
			if len(indexes) == count {
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
	if chunk.ChunkCount <= 0 || chunk.ChunkIndex < 0 || chunk.ChunkIndex >= chunk.ChunkCount {
		s.deliverChunk(chunk.Nonce, memberChunk{err: errors.New("Micup member response has an invalid chunk position; member list is incomplete")})
		return
	}
	response := memberChunk{members: chunk.Members, guildID: chunk.GuildID, index: chunk.ChunkIndex, count: chunk.ChunkCount, last: chunk.ChunkIndex == chunk.ChunkCount-1}
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
	const maxDuration = time.Duration(1<<63 - 1)
	nanoseconds := limited.RetryAfter * float64(time.Second)
	var retryAfter time.Duration
	if nanoseconds >= float64(maxDuration) {
		retryAfter = maxDuration
	} else if nanoseconds > 0 {
		retryAfter = time.Duration(nanoseconds)
	}
	s.deliverChunk(limited.Meta.Nonce, memberChunk{err: &MembersRateLimitedError{GuildID: limited.Meta.GuildID, RetryAfter: retryAfter}})
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
