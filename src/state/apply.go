package state

import (
	"encoding/json"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

type guildPayload struct {
	ID          string             `json:"id"`
	VoiceStates []model.VoiceState `json:"voice_states"`
	Channels    []model.Channel    `json:"channels"`
}

func (c *Cache) Apply(name string, data json.RawMessage) *VoiceMove {
	switch name {
	case gateway.EventReady:
		c.applyReady(data)
	case gateway.EventGuildCreate:
		c.applyGuildCreate(data)
	case gateway.EventGuildDelete:
		c.applyGuildDelete(data)
	case gateway.EventChannelCreate, gateway.EventChannelUpdate, gateway.EventChannelDelete:
		c.applyChannel(name, data)
	case gateway.EventVoiceStateUpdate:
		return c.applyVoiceState(data)
	}
	return nil
}

func (c *Cache) applyReady(data json.RawMessage) {
	var ready struct {
		Guilds []json.RawMessage `json:"guilds"`
	}
	_ = json.Unmarshal(data, &ready)
	c.mu.Lock()
	c.voices = map[string]map[string]string{}
	c.channels = map[string]string{}
	c.mu.Unlock()
	for _, guild := range ready.Guilds {
		c.applyGuildCreate(guild)
	}
}

func (c *Cache) applyGuildCreate(data json.RawMessage) {
	var guild guildPayload
	if json.Unmarshal(data, &guild) != nil || guild.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.voices[guild.ID] = map[string]string{}
	for _, channel := range guild.Channels {
		if channel.ID != "" {
			c.channels[channel.ID] = guild.ID
		}
	}
	for _, voice := range guild.VoiceStates {
		if voice.UserID != "" && voice.ChannelID != "" {
			c.voices[guild.ID][voice.UserID] = voice.ChannelID
		}
	}
}

func (c *Cache) applyGuildDelete(data json.RawMessage) {
	var guild struct {
		ID          string `json:"id"`
		Unavailable bool   `json:"unavailable"`
	}
	if json.Unmarshal(data, &guild) != nil || guild.ID == "" || guild.Unavailable {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.voices, guild.ID)
	for channel, owner := range c.channels {
		if owner == guild.ID {
			delete(c.channels, channel)
		}
	}
}

func (c *Cache) applyChannel(name string, data json.RawMessage) {
	var channel model.Channel
	if json.Unmarshal(data, &channel) != nil || channel.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if name == gateway.EventChannelDelete {
		delete(c.channels, channel.ID)
	} else if channel.GuildID != "" {
		c.channels[channel.ID] = channel.GuildID
	}
}

func (c *Cache) applyVoiceState(data json.RawMessage) *VoiceMove {
	var voice model.VoiceState
	if json.Unmarshal(data, &voice) != nil || voice.GuildID == "" || voice.UserID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.voices[voice.GuildID] == nil {
		c.voices[voice.GuildID] = map[string]string{}
	}
	previous := c.voices[voice.GuildID][voice.UserID]
	if voice.ChannelID == "" {
		delete(c.voices[voice.GuildID], voice.UserID)
	} else {
		c.voices[voice.GuildID][voice.UserID] = voice.ChannelID
	}
	if voice.UserID != c.botID {
		return nil
	}
	return &VoiceMove{GuildID: voice.GuildID, Previous: previous, Current: voice.ChannelID}
}
