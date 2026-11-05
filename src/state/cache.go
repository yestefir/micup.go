package state

import (
	"sync"
)

type Cache struct {
	mu       sync.RWMutex
	botID    string
	voices   map[string]map[string]string
	channels map[string]string
}

type VoiceMove struct {
	GuildID  string
	Previous string
	Current  string
}

func New(botID string) *Cache {
	return &Cache{botID: botID, voices: map[string]map[string]string{}, channels: map[string]string{}}
}

func (c *Cache) GuildForChannel(channelID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	guild, ok := c.channels[channelID]
	return guild, ok
}

func (c *Cache) VoiceChannel(guildID, userID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.voices[guildID][userID]
}

func (c *Cache) BotVoiceChannel(guildID string) string { return c.VoiceChannel(guildID, c.botID) }

func (c *Cache) Guilds() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.voices)
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
	c.channels[channelID] = guildID
	c.mu.Unlock()
}
