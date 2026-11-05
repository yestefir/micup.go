package micup

import (
	"context"
	"errors"
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

func (s *Session) JoinVoice(ctx context.Context, guildID, channelID string, options VoiceJoin) (VoiceServerUpdate, error) {
	if guildID == "" || channelID == "" {
		return VoiceServerUpdate{}, errors.New("guild ID and voice channel ID are required")
	}
	if !s.Ready() {
		return VoiceServerUpdate{}, ErrNotReady
	}
	s.mu.Lock()
	if s.pending[guildID] != nil {
		s.mu.Unlock()
		return VoiceServerUpdate{}, ErrVoiceJoinPending
	}
	pending := make(chan VoiceServerUpdate, 1)
	s.pending[guildID] = pending
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.pending, guildID)
		s.mu.Unlock()
	}()
	if err := s.Send(OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": channelID, "self_mute": options.SelfMute, "self_deaf": options.SelfDeaf, "self_video": false, "self_stream": false}); err != nil {
		return VoiceServerUpdate{}, err
	}
	select {
	case <-ctx.Done():
		return VoiceServerUpdate{}, ctx.Err()
	case grant := <-pending:
		if grant.Endpoint == "" || grant.Token == "" {
			return VoiceServerUpdate{}, ErrVoiceGrant
		}
		if grant.ChannelID == "" {
			grant.ChannelID = channelID
		}
		if grant.ChannelID != channelID {
			return VoiceServerUpdate{}, ErrVoiceMoved
		}
		s.mu.Lock()
		s.grants[guildID] = grant
		s.mu.Unlock()
		return grant, nil
	}
}

func (s *Session) Voice(guildID string) (VoiceServerUpdate, bool) {
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
	return s.Send(OpVoiceStateUpdate, map[string]any{"guild_id": guildID, "channel_id": nil, "connection_id": connectionID})
}

func (s *Session) handleBotVoiceMove(move *botVoiceMove) {
	if move == nil || move.previous == "" || move.previous == move.current {
		return
	}
	s.mu.Lock()
	grant := s.grants[move.guildID]
	delete(s.grants, move.guildID)
	s.mu.Unlock()
	s.voiceLost(move.guildID, grant.ConnectionID)
}

func (s *Session) handleVoiceServerUpdate(event Event) {
	var grant VoiceServerUpdate
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
