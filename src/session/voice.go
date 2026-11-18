package session

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/state"
)

var (
	ErrVoiceJoinPending = errors.New("Micup voice join is already pending")
	ErrVoiceGrant       = errors.New("invalid Micup voice grant")
	ErrVoiceMoved       = errors.New("Micup voice channel changed during join")
	ErrVoiceRefused     = errors.New("Micup refused the voice join")
	ErrNotInVoice       = errors.New("the bot has no voice connection in the guild")
)

type VoiceJoin struct {
	SelfMute bool
	SelfDeaf bool
}

type VoiceRefusedError struct {
	GuildID string
	Code    string
	Message string
}

func (e *VoiceRefusedError) Error() string {
	return fmt.Sprintf("Micup refused the voice join for guild %s (%s)", e.GuildID, e.Code)
}

func (e *VoiceRefusedError) Is(target error) bool { return target == ErrVoiceRefused }

type joinResult struct {
	grant model.VoiceServerUpdate
	err   error
}

type voicePending struct {
	results chan joinResult
	closed  bool
}

func (s *Session) JoinVoice(ctx context.Context, guildID, channelID string, options VoiceJoin) (model.VoiceServerUpdate, error) {
	if guildID == "" || channelID == "" {
		return model.VoiceServerUpdate{}, errors.New("guild ID and voice channel ID are required")
	}
	if !s.Ready() {
		return model.VoiceServerUpdate{}, ErrNotReady
	}
	if err := s.checkVoicePermissions(guildID, channelID); err != nil {
		return model.VoiceServerUpdate{}, err
	}
	pending, err := s.beginJoin(guildID)
	if err != nil {
		return model.VoiceServerUpdate{}, err
	}
	defer s.endJoin(guildID, pending)
	if err := s.Send(ctx, gateway.OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": channelID, "self_mute": options.SelfMute, "self_deaf": options.SelfDeaf, "self_video": false, "self_stream": false}); err != nil {
		return model.VoiceServerUpdate{}, err
	}
	select {
	case <-ctx.Done():
		return model.VoiceServerUpdate{}, ctx.Err()
	case result := <-pending:
		return s.finishJoin(guildID, channelID, pending, result)
	}
}

func (s *Session) finishJoin(guildID, channelID string, pending chan joinResult, result joinResult) (model.VoiceServerUpdate, error) {
	if result.err != nil {
		return model.VoiceServerUpdate{}, result.err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.pending[guildID]
	if current.results != pending {
		return model.VoiceServerUpdate{}, ErrVoiceGrant
	}
	if current.closed {
		return model.VoiceServerUpdate{}, gateway.ErrClosed
	}
	// A newer grant may have arrived after the caller received its result.
	// Select and validate it under the same lock used by delivery and commit.
	select {
	case result = <-pending:
	default:
	}
	if result.err != nil {
		return model.VoiceServerUpdate{}, result.err
	}
	grant := result.grant
	if grant.Endpoint == "" || grant.Token == "" {
		return model.VoiceServerUpdate{}, ErrVoiceGrant
	}
	if grant.ChannelID == "" {
		grant.ChannelID = channelID
	}
	if grant.ChannelID != channelID {
		return model.VoiceServerUpdate{}, ErrVoiceMoved
	}
	s.grants[guildID] = grant
	// Further grants now belong to the established connection, even if the
	// caller has not yet run its deferred cleanup.
	delete(s.pending, guildID)
	s.joinOrder = slices.DeleteFunc(s.joinOrder, func(id string) bool { return id == guildID })
	return grant, nil
}

func (s *Session) Voice(guildID string) (model.VoiceServerUpdate, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	grant, ok := s.grants[guildID]
	return grant, ok
}

func (s *Session) LeaveVoice(guildID, connectionID string) error {
	s.mu.Lock()
	grant, ok := s.grants[guildID]
	if !ok || connectionID == "" || grant.ConnectionID != connectionID {
		s.mu.Unlock()
		return nil
	}
	delete(s.grants, guildID)
	s.mu.Unlock()
	return s.Send(context.Background(), gateway.OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": nil, "connection_id": connectionID})
}

func (s *Session) UpdateVoiceState(ctx context.Context, guildID string, options VoiceJoin) error {
	s.mu.Lock()
	grant, ok := s.grants[guildID]
	s.mu.Unlock()
	if !ok || grant.ConnectionID == "" || grant.ChannelID == "" {
		return ErrNotInVoice
	}
	return s.Send(ctx, gateway.OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": grant.ChannelID, "connection_id": grant.ConnectionID, "self_mute": options.SelfMute, "self_deaf": options.SelfDeaf, "self_video": false, "self_stream": false})
}

func (s *Session) beginJoin(guildID string) (chan joinResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending[guildID].results != nil {
		return nil, ErrVoiceJoinPending
	}
	pending := make(chan joinResult, 1)
	s.pending[guildID] = voicePending{results: pending}
	s.joinOrder = append(s.joinOrder, guildID)
	return pending, nil
}

func (s *Session) endJoin(guildID string, pending chan joinResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending[guildID].results != pending {
		return
	}
	delete(s.pending, guildID)
	s.joinOrder = slices.DeleteFunc(s.joinOrder, func(id string) bool { return id == guildID })
}

func (s *Session) handleGatewayError(refusal gateway.GatewayError) {
	s.mu.Lock()
	if len(s.joinOrder) == 0 {
		s.mu.Unlock()
		return
	}
	guildID := s.joinOrder[0]
	s.joinOrder = s.joinOrder[1:]
	pending := s.pending[guildID].results
	if pending != nil {
		select {
		case pending <- joinResult{err: &VoiceRefusedError{GuildID: guildID, Code: refusal.Code, Message: refusal.Message}}:
		default:
		}
	}
	s.mu.Unlock()
}

func (s *Session) handleBotVoiceMove(move *state.VoiceMove) {
	if move == nil || move.Previous == "" || move.Previous == move.Current {
		return
	}
	s.mu.Lock()
	grant, had := s.grants[move.GuildID]
	if had && move.Current != "" && grant.ChannelID == move.Current {
		s.mu.Unlock()
		return
	}
	delete(s.grants, move.GuildID)
	s.mu.Unlock()
	s.voiceLost(move.GuildID, grant.ConnectionID)
}

func (s *Session) handleVoiceServerUpdate(event gateway.Event) {
	var grant model.VoiceServerUpdate
	if event.Decode(&grant) != nil || grant.GuildID == "" {
		return
	}
	s.mu.Lock()
	entry := s.pending[grant.GuildID]
	pending := entry.results
	if pending != nil {
		if !entry.closed {
			s.joinOrder = slices.DeleteFunc(s.joinOrder, func(id string) bool { return id == grant.GuildID })
			result := joinResult{grant: grant}
			select {
			case previous := <-pending:
				if previous.err != nil {
					result = previous
				}
			default:
			}
			pending <- result
		}
		s.mu.Unlock()
		return
	}
	previous, had := s.grants[grant.GuildID]
	rotated := s.options.OnVoiceGrant != nil && grant.Endpoint != "" && grant.Token != ""
	if rotated {
		s.grants[grant.GuildID] = grant
	} else {
		delete(s.grants, grant.GuildID)
	}
	s.mu.Unlock()
	switch {
	case rotated:
		s.options.OnVoiceGrant(grant)
	case had:
		s.voiceLost(grant.GuildID, previous.ConnectionID)
	}
}

func (s *Session) stopVoiceJoins() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failVoiceJoins(gateway.ErrClosed)
}

// failVoiceJoins requires s.mu. Callers decide whether reservations remain
// until caller cleanup (shutdown), or are invalidated (a replaced session).
func (s *Session) failVoiceJoins(err error) {
	for guildID, pending := range s.pending {
		// Keep ownership until caller cleanup, but invalidate even a result
		// the caller already consumed before this notification.
		pending.closed = true
		s.pending[guildID] = pending
		select {
		case <-pending.results:
		default:
		}
		pending.results <- joinResult{err: err}
	}
	s.joinOrder = nil
}

func (s *Session) dropAllGrants() {
	s.mu.Lock()
	grants := s.grants
	s.grants = map[string]model.VoiceServerUpdate{}
	s.failVoiceJoins(ErrVoiceGrant)
	clear(s.pending)
	s.mu.Unlock()
	for guildID, grant := range grants {
		s.voiceLost(guildID, grant.ConnectionID)
	}
}

func (s *Session) voiceLost(guildID, connectionID string) {
	if s.options.OnVoiceLost != nil {
		s.options.OnVoiceLost(guildID, connectionID)
	}
}
