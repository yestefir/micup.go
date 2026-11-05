package micup

import (
	"encoding/json"
	"sync"
)

type State struct {
	mu       sync.RWMutex
	botID    string
	voices   map[string]map[string]string
	channels map[string]string
}

func NewState(botID string) *State {
	return &State{botID: botID, voices: map[string]map[string]string{}, channels: map[string]string{}}
}

func (s *State) GuildForChannel(channelID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	guild, ok := s.channels[channelID]
	return guild, ok
}

func (s *State) VoiceChannel(guildID, userID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.voices[guildID][userID]
}

func (s *State) BotVoiceChannel(guildID string) string { return s.VoiceChannel(guildID, s.botID) }

func (s *State) Guilds() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.voices)
}

func (s *State) Channels() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.channels)
}

func (s *State) rememberChannel(channelID, guildID string) {
	if channelID == "" || guildID == "" {
		return
	}
	s.mu.Lock()
	s.channels[channelID] = guildID
	s.mu.Unlock()
}

type botVoiceMove struct {
	guildID  string
	previous string
	current  string
}

type gatewayGuild struct {
	ID          string       `json:"id"`
	VoiceStates []VoiceState `json:"voice_states"`
	Channels    []Channel    `json:"channels"`
}

func (s *State) Apply(event Event) {
	s.apply(event.Name, event.Data)
}

func (s *State) apply(name string, data json.RawMessage) *botVoiceMove {
	switch name {
	case EventReady:
		var ready struct {
			Guilds []json.RawMessage `json:"guilds"`
		}
		_ = json.Unmarshal(data, &ready)
		s.mu.Lock()
		s.voices = map[string]map[string]string{}
		s.channels = map[string]string{}
		s.mu.Unlock()
		for _, guild := range ready.Guilds {
			s.apply(EventGuildCreate, guild)
		}
	case EventGuildCreate:
		var guild gatewayGuild
		if json.Unmarshal(data, &guild) != nil || guild.ID == "" {
			return nil
		}
		s.mu.Lock()
		s.voices[guild.ID] = map[string]string{}
		for _, channel := range guild.Channels {
			if channel.ID != "" {
				s.channels[channel.ID] = guild.ID
			}
		}
		for _, voice := range guild.VoiceStates {
			if voice.UserID != "" && voice.ChannelID != "" {
				s.voices[guild.ID][voice.UserID] = voice.ChannelID
			}
		}
		s.mu.Unlock()
	case EventGuildDelete:
		var guild struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(data, &guild) != nil || guild.ID == "" {
			return nil
		}
		s.mu.Lock()
		delete(s.voices, guild.ID)
		for channel, owner := range s.channels {
			if owner == guild.ID {
				delete(s.channels, channel)
			}
		}
		s.mu.Unlock()
	case EventChannelCreate, EventChannelUpdate, EventChannelDelete:
		var channel Channel
		if json.Unmarshal(data, &channel) != nil || channel.ID == "" {
			return nil
		}
		s.mu.Lock()
		if name == EventChannelDelete {
			delete(s.channels, channel.ID)
		} else if channel.GuildID != "" {
			s.channels[channel.ID] = channel.GuildID
		}
		s.mu.Unlock()
	case EventVoiceStateUpdate:
		var voice VoiceState
		if json.Unmarshal(data, &voice) != nil || voice.GuildID == "" || voice.UserID == "" {
			return nil
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.voices[voice.GuildID] == nil {
			s.voices[voice.GuildID] = map[string]string{}
		}
		previous := s.voices[voice.GuildID][voice.UserID]
		if voice.ChannelID == "" {
			delete(s.voices[voice.GuildID], voice.UserID)
		} else {
			s.voices[voice.GuildID][voice.UserID] = voice.ChannelID
		}
		if voice.UserID == s.botID {
			return &botVoiceMove{guildID: voice.GuildID, previous: previous, current: voice.ChannelID}
		}
	}
	return nil
}
