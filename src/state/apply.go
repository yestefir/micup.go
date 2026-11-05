package state

import (
	"encoding/json"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

type guildPayload struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable"`
	Properties  struct {
		OwnerID string `json:"owner_id"`
	} `json:"properties"`
	OwnerID     string              `json:"owner_id"`
	Roles       []model.Role        `json:"roles"`
	Channels    []model.Channel     `json:"channels"`
	Members     []model.Member      `json:"members"`
	VoiceStates []voiceStatePayload `json:"voice_states"`
}

type voiceStatePayload struct {
	model.VoiceState
	Member *model.Member `json:"member"`
}

func (c *Cache) Apply(name string, data json.RawMessage) *VoiceMove {
	switch name {
	case gateway.EventReady:
		c.applyReady(data)
	case gateway.EventGuildCreate, gateway.EventGuildSync:
		c.applyGuild(data)
	case gateway.EventGuildUpdate:
		c.applyGuildUpdate(data)
	case gateway.EventGuildDelete:
		c.applyGuildDelete(data)
	case gateway.EventGuildRoleCreate, gateway.EventGuildRoleUpdate:
		c.applyRole(data)
	case gateway.EventGuildRoleUpdateBulk:
		c.applyRoles(data)
	case gateway.EventGuildRoleDelete:
		c.applyRoleDelete(data)
	case gateway.EventGuildMemberAdd, gateway.EventGuildMemberUpdate:
		c.applyMember(data)
	case gateway.EventGuildMemberRemove:
		c.applyMemberRemove(data)
	case gateway.EventChannelCreate, gateway.EventChannelUpdate, gateway.EventChannelDelete:
		c.applyChannel(name, data)
	case gateway.EventChannelUpdateBulk:
		c.applyChannels(data)
	case gateway.EventMessageCreate:
		c.applyMessage(data)
	case gateway.EventGuildMembersChunk:
		c.applyMembersChunk(data)
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
	c.guilds = map[string]*guild{}
	c.channels = map[string]model.Channel{}
	c.mu.Unlock()
	for _, g := range ready.Guilds {
		c.applyGuild(g)
	}
}

func (c *Cache) applyGuild(data json.RawMessage) {
	var payload guildPayload
	if json.Unmarshal(data, &payload) != nil || payload.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if payload.Unavailable {
		if g := c.guilds[payload.ID]; g != nil {
			g.unavailable = true
		} else {
			g = newGuild()
			g.unavailable = true
			c.guilds[payload.ID] = g
		}
		return
	}
	previous := c.guilds[payload.ID]
	g := newGuild()
	if previous != nil {
		g.members = previous.members
	}
	g.ownerID = payload.Properties.OwnerID
	if g.ownerID == "" {
		g.ownerID = payload.OwnerID
	}
	for _, role := range payload.Roles {
		g.roles[role.ID] = role
	}
	for id, channel := range c.channels {
		if channel.GuildID == payload.ID {
			delete(c.channels, id)
		}
	}
	for _, channel := range payload.Channels {
		if channel.ID != "" {
			channel.GuildID = payload.ID
			c.channels[channel.ID] = channel
		}
	}
	for _, member := range payload.Members {
		if member.User != nil && member.User.ID != "" {
			g.members[member.User.ID] = member.Roles
		}
	}
	for _, voice := range payload.VoiceStates {
		if voice.UserID != "" && voice.ChannelID != "" {
			g.voices[voice.UserID] = voice.ChannelID
		}
		if voice.Member != nil && voice.UserID != "" {
			g.members[voice.UserID] = voice.Member.Roles
		}
	}
	c.guilds[payload.ID] = g
}

func (c *Cache) applyGuildUpdate(data json.RawMessage) {
	var payload struct {
		ID      string `json:"id"`
		GuildID string `json:"guild_id"`
		OwnerID string `json:"owner_id"`
	}
	if json.Unmarshal(data, &payload) != nil {
		return
	}
	id := payload.GuildID
	if id == "" {
		id = payload.ID
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.guilds[id]; g != nil && payload.OwnerID != "" {
		g.ownerID = payload.OwnerID
	}
}

func (c *Cache) applyGuildDelete(data json.RawMessage) {
	var payload struct {
		ID          string `json:"id"`
		Unavailable bool   `json:"unavailable"`
	}
	if json.Unmarshal(data, &payload) != nil || payload.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if payload.Unavailable {
		if g := c.guilds[payload.ID]; g != nil {
			g.unavailable = true
		}
		return
	}
	delete(c.guilds, payload.ID)
	for id, channel := range c.channels {
		if channel.GuildID == payload.ID {
			delete(c.channels, id)
		}
	}
}

func (c *Cache) guildLocked(id string) *guild {
	g := c.guilds[id]
	if g == nil {
		g = newGuild()
		c.guilds[id] = g
	}
	return g
}

func (c *Cache) applyRole(data json.RawMessage) {
	var payload struct {
		GuildID string     `json:"guild_id"`
		Role    model.Role `json:"role"`
	}
	if json.Unmarshal(data, &payload) != nil || payload.GuildID == "" || payload.Role.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.guildLocked(payload.GuildID).roles[payload.Role.ID] = payload.Role
}

func (c *Cache) applyRoles(data json.RawMessage) {
	var payload struct {
		GuildID string       `json:"guild_id"`
		Roles   []model.Role `json:"roles"`
	}
	if json.Unmarshal(data, &payload) != nil || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	g := c.guildLocked(payload.GuildID)
	for _, role := range payload.Roles {
		if role.ID != "" {
			g.roles[role.ID] = role
		}
	}
}

func (c *Cache) applyRoleDelete(data json.RawMessage) {
	var payload struct {
		GuildID string `json:"guild_id"`
		RoleID  string `json:"role_id"`
	}
	if json.Unmarshal(data, &payload) != nil || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.guilds[payload.GuildID]; g != nil {
		delete(g.roles, payload.RoleID)
	}
}

func (c *Cache) applyMember(data json.RawMessage) {
	var payload model.GuildMember
	if json.Unmarshal(data, &payload) != nil || payload.GuildID == "" || payload.User == nil || payload.User.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.guildLocked(payload.GuildID).members[payload.User.ID] = payload.Roles
}

func (c *Cache) applyMemberRemove(data json.RawMessage) {
	var payload struct {
		GuildID string `json:"guild_id"`
		User    struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if json.Unmarshal(data, &payload) != nil || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.guilds[payload.GuildID]; g != nil {
		delete(g.members, payload.User.ID)
		delete(g.voices, payload.User.ID)
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
		c.channels[channel.ID] = channel
	}
}

func (c *Cache) applyChannels(data json.RawMessage) {
	var payload struct {
		GuildID  string          `json:"guild_id"`
		Channels []model.Channel `json:"channels"`
	}
	if json.Unmarshal(data, &payload) != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, channel := range payload.Channels {
		if channel.ID == "" {
			continue
		}
		if channel.GuildID == "" {
			channel.GuildID = payload.GuildID
		}
		c.channels[channel.ID] = channel
	}
}

func (c *Cache) applyMessage(data json.RawMessage) {
	var message struct {
		GuildID string `json:"guild_id"`
		Author  struct {
			ID string `json:"id"`
		} `json:"author"`
		Member *model.Member `json:"member"`
	}
	if json.Unmarshal(data, &message) != nil || message.GuildID == "" || message.Member == nil || message.Author.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.guilds[message.GuildID]; g != nil {
		g.members[message.Author.ID] = message.Member.Roles
	}
}

func (c *Cache) applyVoiceState(data json.RawMessage) *VoiceMove {
	var voice voiceStatePayload
	if json.Unmarshal(data, &voice) != nil || voice.GuildID == "" || voice.UserID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	g := c.guildLocked(voice.GuildID)
	previous := g.voices[voice.UserID]
	if voice.ChannelID == "" {
		delete(g.voices, voice.UserID)
	} else {
		g.voices[voice.UserID] = voice.ChannelID
	}
	if voice.Member != nil {
		g.members[voice.UserID] = voice.Member.Roles
	}
	if voice.UserID != c.botID {
		return nil
	}
	return &VoiceMove{GuildID: voice.GuildID, Previous: previous, Current: voice.ChannelID}
}

func (c *Cache) applyMembersChunk(data json.RawMessage) {
	var chunk struct {
		GuildID string         `json:"guild_id"`
		Members []model.Member `json:"members"`
	}
	if json.Unmarshal(data, &chunk) != nil || chunk.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	g := c.guilds[chunk.GuildID]
	if g == nil {
		return
	}
	for _, member := range chunk.Members {
		if member.User != nil && member.User.ID != "" {
			g.members[member.User.ID] = member.Roles
		}
	}
}
