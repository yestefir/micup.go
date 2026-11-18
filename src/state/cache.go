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
	info        model.Guild
	emojis      []model.Emoji
	stickers    []model.Sticker
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
		if channel.IsThread() {
			parent, ok := c.channels[channel.ParentID]
			if !ok {
				return 0, false
			}
			channel = parent
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

func (c *Cache) Guild(guildID string) (model.Guild, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil || g.unavailable {
		return model.Guild{}, false
	}
	guild := g.info
	guild.ID = guildID
	guild.OwnerID = g.ownerID
	guild.Features = slices.Clone(guild.Features)
	guild.Roles = sortedRoles(g.roles)
	guild.Emojis = slices.Clone(g.emojis)
	guild.Stickers = slices.Clone(g.stickers)
	return guild, true
}

func (c *Cache) GuildIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids := make([]string, 0, len(c.guilds))
	for id := range c.guilds {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, model.CompareSnowflakes)
	return ids
}

func (c *Cache) GuildChannels(guildID string) []model.Channel {
	return c.selectChannels(func(channel model.Channel) bool { return channel.GuildID == guildID && !channel.IsThread() })
}

func (c *Cache) Threads(guildID string) []model.Channel {
	return c.selectChannels(func(channel model.Channel) bool { return channel.GuildID == guildID && channel.IsThread() })
}

func (c *Cache) Roles(guildID string) []model.Role {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil {
		return nil
	}
	return sortedRoles(g.roles)
}

func sortedRoles(byID map[string]model.Role) []model.Role {
	roles := make([]model.Role, 0, len(byID))
	for _, role := range byID {
		roles = append(roles, role)
	}
	slices.SortFunc(roles, func(a, b model.Role) int {
		if a.Position != b.Position {
			return b.Position - a.Position
		}
		return model.CompareSnowflakes(a.ID, b.ID)
	})
	return roles
}

func (c *Cache) Emojis(guildID string) []model.Emoji {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if g := c.guilds[guildID]; g != nil {
		return slices.Clone(g.emojis)
	}
	return nil
}

func (c *Cache) Stickers(guildID string) []model.Sticker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if g := c.guilds[guildID]; g != nil {
		return slices.Clone(g.stickers)
	}
	return nil
}

func (c *Cache) Members(guildID string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if g := c.guilds[guildID]; g != nil {
		return len(g.members)
	}
	return 0
}

func (c *Cache) HighestRole(guildID, userID string) (model.Role, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil {
		return model.Role{}, false
	}
	return g.highestRole(guildID, userID)
}

func (c *Cache) CanManage(guildID, actorID, targetID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g := c.guilds[guildID]
	if g == nil {
		return false
	}
	if g.ownerID != "" && actorID == g.ownerID {
		return true
	}
	if targetID == g.ownerID {
		return false
	}
	actor, ok := g.highestRole(guildID, actorID)
	if !ok {
		return false
	}
	target, ok := g.highestRole(guildID, targetID)
	return ok && actor.Position > target.Position
}

func (g *guild) highestRole(guildID, userID string) (model.Role, bool) {
	best, found := g.roles[guildID]
	for _, id := range g.members[userID] {
		if role, ok := g.roles[id]; ok && (!found || role.Position > best.Position) {
			best, found = role, true
		}
	}
	return best, found
}

func (c *Cache) selectChannels(keep func(model.Channel) bool) []model.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var channels []model.Channel
	for _, channel := range c.channels {
		if keep(channel) {
			channel.PermissionOverwrites = slices.Clone(channel.PermissionOverwrites)
			channels = append(channels, channel)
		}
	}
	slices.SortFunc(channels, func(a, b model.Channel) int {
		if a.Position != b.Position {
			return a.Position - b.Position
		}
		return model.CompareSnowflakes(a.ID, b.ID)
	})
	return channels
}
