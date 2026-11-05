package session

import (
	"context"
	"errors"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/state"
)

var (
	ErrVoiceJoinPending = errors.New("Micup voice join is already pending")
	ErrVoiceGrant       = errors.New("invalid Micup voice grant")
	ErrVoiceMoved       = errors.New("Micup voice channel changed during join")
)

type VoiceJoin struct {
	SelfMute bool
	SelfDeaf bool
}

func (s *Session) JoinVoice(ctx context.Context, guildID, channelID string, options VoiceJoin) (model.VoiceServerUpdate, error) {
	if guildID == "" || channelID == "" {
		return model.VoiceServerUpdate{}, errors.New("guild ID and voice channel ID are required")
	}
	if !s.Ready() {
		return model.VoiceServerUpdate{}, ErrNotReady
	}
	pending, err := s.beginJoin(guildID)
	if err != nil {
		return model.VoiceServerUpdate{}, err
	}
	defer s.endJoin(guildID)
	if err := s.Send(gateway.OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": channelID, "self_mute": options.SelfMute, "self_deaf": options.SelfDeaf, "self_video": false, "self_stream": false}); err != nil {
		return model.VoiceServerUpdate{}, err
	}
	select {
	case <-ctx.Done():
		return model.VoiceServerUpdate{}, ctx.Err()
	case grant := <-pending:
		if grant.Endpoint == "" || grant.Token == "" {
			return model.VoiceServerUpdate{}, ErrVoiceGrant
		}
		if grant.ChannelID == "" {
			grant.ChannelID = channelID
		}
		if grant.ChannelID != channelID {
			return model.VoiceServerUpdate{}, ErrVoiceMoved
		}
		s.mu.Lock()
		s.grants[guildID] = grant
		s.mu.Unlock()
		return grant, nil
	}
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
	return s.Send(gateway.OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": nil, "connection_id": connectionID})
}

func (s *Session) beginJoin(guildID string) (chan model.VoiceServerUpdate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending[guildID] != nil {
		return nil, ErrVoiceJoinPending
	}
	pending := make(chan model.VoiceServerUpdate, 1)
	s.pending[guildID] = pending
	return pending, nil
}

func (s *Session) endJoin(guildID string) {
	s.mu.Lock()
	delete(s.pending, guildID)
	s.mu.Unlock()
}

func (s *Session) handleBotVoiceMove(move *state.VoiceMove) {
	if move == nil || move.Previous == "" || move.Previous == move.Current {
		return
	}
	s.mu.Lock()
	grant := s.grants[move.GuildID]
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
	pending := s.pending[grant.GuildID]
	previous, had := s.grants[grant.GuildID]
	if pending == nil {
		delete(s.grants, grant.GuildID)
	}
	s.mu.Unlock()
	if pending != nil {
		select {
		case pending <- grant:
		default:
		}
	} else if had {
		s.voiceLost(grant.GuildID, previous.ConnectionID)
	}
}

func (s *Session) voiceLost(guildID, connectionID string) {
	if s.options.OnVoiceLost != nil {
		s.options.OnVoiceLost(guildID, connectionID)
	}
}
