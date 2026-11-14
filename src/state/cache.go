package state

import (
	"slices"
	"sync"

	"github.com/yestefir/micup.go/src/model"
)

type Cache struct {
	mu       sync.RWMutex
	botID    string
	guilds   map[string]*guild
	channels map[string]model.Channel
}

type guild struct {
	ownerID     string
	unavailable bool
	roles       map[string]model.Role
	members     map[string][]string
	voices      map[string]string
}

type VoiceMove struct {
	GuildID  string
	Previous string
	Current  string
}

func New(botID string) *Cache {
	return &Cache{botID: botID, guilds: map[string]*guild{}, channels: map[string]model.Channel{}}
}

func newGuild() *guild {
	return &guild{roles: map[string]model.Role{}, members: map[string][]string{}, voices: map[string]string{}}
}

func (c *Cache) BotID() string { return c.botID }

func (c *Cache) GuildForChannel(channelID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channel, ok := c.channels[channelID]
	if !ok || channel.GuildID == "" {
		return "", false
	}
	return channel.GuildID, true
}

func (c *Cache) Channel(channelID string) (model.Channel, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channel, ok := c.channels[channelID]
	if ok {
		copyChannel(&channel)
	}
	return channel, ok
}

func (c *Cache) Role(guildID, roleID string) (model.Role, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil {
		return model.Role{}, false
	}
	role, ok := g.roles[roleID]
	copyRole(&role)
	return role, ok
}

func (c *Cache) MemberRoles(guildID, userID string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil {
		return nil, false
	}
	roles, ok := g.members[userID]
	return slices.Clone(roles), ok
}

func (c *Cache) OwnerID(guildID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if g := c.guilds[guildID]; g != nil {
		return g.ownerID
	}
	return ""
}

func (c *Cache) Unavailable(guildID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	return g != nil && g.unavailable
}

func (c *Cache) Permissions(guildID, channelID, userID string) (model.Permissions, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil || len(g.roles) == 0 {
		return 0, false
	}
	if _, ok := g.roles[guildID]; !ok {
		return 0, false
	}
	memberRoles, isMember := g.members[userID]
	if !isMember && userID != g.ownerID {
		return 0, false
	}
	var overwrites []model.PermissionOverwrite
	if channelID != "" {
		channel, ok := c.channels[channelID]
		if !ok || channel.GuildID != guildID {
			return 0, false
		}
		overwrites = channel.PermissionOverwrites
	}
	return model.ComputePermissions(guildID, g.ownerID, userID, g.roles, memberRoles, overwrites), true
}

func (c *Cache) BotPermissions(guildID, channelID string) (model.Permissions, bool) {
	return c.Permissions(guildID, channelID, c.botID)
}

func (c *Cache) VoiceChannel(guildID, userID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if g := c.guilds[guildID]; g != nil {
		return g.voices[userID]
	}
	return ""
}

func (c *Cache) VoiceMembers(guildID, channelID string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil {
		return nil
	}
	var users []string
	for user, channel := range g.voices {
		if channel == channelID {
			users = append(users, user)
		}
	}
	slices.Sort(users)
	return users
}

func (c *Cache) BotVoiceChannel(guildID string) string { return c.VoiceChannel(guildID, c.botID) }

func (c *Cache) Guilds() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.guilds)
}

func (c *Cache) Channels() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.channels)
}

func (c *Cache) Remember(channelID, guildID string) {
	if channelID == "" || guildID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if channel, ok := c.channels[channelID]; ok {
		channel.GuildID = guildID
		c.channels[channelID] = channel
		return
	}
	c.channels[channelID] = model.Channel{ID: channelID, GuildID: guildID}
}
