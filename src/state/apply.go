package state

import (
	"encoding/json"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

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
	case gateway.EventGuildEmojisUpdate:
		c.applyEmojis(data)
	case gateway.EventGuildStickersUpdate:
		c.applyStickers(data)
	case gateway.EventThreadCreate, gateway.EventThreadUpdate:
		c.applyThread(data)
	case gateway.EventThreadDelete:
		c.applyThreadDelete(data)
	case gateway.EventThreadListSync:
		c.applyThreadListSync(data)
	}
	return nil
}

func (c *Cache) applyReady(data json.RawMessage) {
	ready, _ := events.Parse[events.Ready](data)
	c.mu.Lock()
	c.guilds = map[string]*guild{}
	c.channels = map[string]model.Channel{}
	c.mu.Unlock()
	for _, g := range ready.Guilds {
		c.applyGuildCreate(g)
	}
}

func (c *Cache) applyGuild(data json.RawMessage) {
	if payload, ok := events.Parse[events.GuildCreate](data); ok {
		c.applyGuildCreate(payload)
	}
}

func (c *Cache) applyGuildCreate(payload events.GuildCreate) {
	if payload.ID == "" {
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
	g.ownerID = payload.OwnerID
	g.info = guildInfo(payload.Guild)
	g.emojis = payload.Emojis
	g.stickers = payload.Stickers
	for _, role := range payload.Roles {
		g.roles[role.ID] = role
	}
	for id, channel := range c.channels {
		if channel.GuildID == payload.ID {
			delete(c.channels, id)
		}
	}
	for _, channel := range append(payload.Channels, payload.Threads...) {
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
	payload, ok := events.Parse[events.GuildUpdate](data)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	g := c.guilds[payload.ID]
	if g == nil {
		return
	}
	if payload.OwnerID != "" {
		g.ownerID = payload.OwnerID
	}
	if payload.Name != "" {
		g.info = guildInfo(payload.Guild)
	}
	for _, role := range payload.Roles {
		if role.ID != "" {
			g.roles[role.ID] = role
		}
	}
}

func (c *Cache) applyGuildDelete(data json.RawMessage) {
	payload, ok := events.Parse[events.GuildDelete](data)
	if !ok || payload.ID == "" {
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
	payload, ok := events.Parse[events.RoleEvent](data)
	if !ok || payload.GuildID == "" || payload.Role.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.guildLocked(payload.GuildID).roles[payload.Role.ID] = payload.Role
}

func (c *Cache) applyRoles(data json.RawMessage) {
	payload, ok := events.Parse[events.RoleUpdateBulk](data)
	if !ok || payload.GuildID == "" {
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
	payload, ok := events.Parse[events.RoleDelete](data)
	if !ok || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.guilds[payload.GuildID]; g != nil {
		delete(g.roles, payload.RoleID)
	}
}

func (c *Cache) applyMember(data json.RawMessage) {
	payload, ok := events.Parse[events.GuildMember](data)
	if !ok || payload.GuildID == "" || payload.User == nil || payload.User.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.guildLocked(payload.GuildID).members[payload.User.ID] = payload.Roles
}

func (c *Cache) applyMemberRemove(data json.RawMessage) {
	payload, ok := events.Parse[events.GuildMemberRemove](data)
	if !ok || payload.GuildID == "" {
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
	channel, ok := events.Parse[model.Channel](data)
	if !ok || channel.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if name == gateway.EventChannelDelete {
		delete(c.channels, channel.ID)
		for id, child := range c.channels {
			if child.IsThread() && child.ParentID == channel.ID {
				delete(c.channels, id)
			}
		}
	} else if channel.GuildID != "" {
		c.channels[channel.ID] = channel
	}
}

func (c *Cache) applyChannels(data json.RawMessage) {
	payload, ok := events.Parse[events.ChannelUpdateBulk](data)
	if !ok {
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
	message, ok := events.Parse[events.MessageCreate](data)
	if !ok || message.GuildID == "" || message.Member == nil || message.Author.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.guilds[message.GuildID]; g != nil {
		g.members[message.Author.ID] = message.Member.Roles
	}
}

func (c *Cache) applyVoiceState(data json.RawMessage) *VoiceMove {
	voice, ok := events.Parse[model.VoiceState](data)
	if !ok || voice.GuildID == "" || voice.UserID == "" {
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
	chunk, ok := events.Parse[events.MembersChunk](data)
	if !ok || chunk.GuildID == "" {
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

func guildInfo(guild model.Guild) model.Guild {
	guild.Roles, guild.Emojis, guild.Stickers = nil, nil, nil
	return guild
}

func (c *Cache) applyEmojis(data json.RawMessage) {
	payload, ok := events.Parse[events.GuildEmojisUpdate](data)
	if !ok || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.guildLocked(payload.GuildID).emojis = payload.Emojis
}

func (c *Cache) applyStickers(data json.RawMessage) {
	payload, ok := events.Parse[events.GuildStickersUpdate](data)
	if !ok || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.guildLocked(payload.GuildID).stickers = payload.Stickers
}

func (c *Cache) applyThread(data json.RawMessage) {
	thread, ok := events.Parse[model.Channel](data)
	if !ok || thread.ID == "" || thread.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels[thread.ID] = thread
}

func (c *Cache) applyThreadDelete(data json.RawMessage) {
	payload, ok := events.Parse[events.ThreadDelete](data)
	if !ok || payload.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.channels, payload.ID)
}

func (c *Cache) applyThreadListSync(data json.RawMessage) {
	payload, ok := events.Parse[events.ThreadListSync](data)
	if !ok || payload.GuildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	parents := map[string]bool{}
	for _, id := range payload.ChannelIDs {
		parents[id] = true
	}
	for id, channel := range c.channels {
		if channel.GuildID == payload.GuildID && channel.IsThread() && (len(parents) == 0 || parents[channel.ParentID]) {
			delete(c.channels, id)
		}
	}
	for _, thread := range payload.Threads {
		if thread.ID != "" {
			thread.GuildID = payload.GuildID
			c.channels[thread.ID] = thread
		}
	}
}
