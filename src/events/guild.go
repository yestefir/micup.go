package events

import (
	"encoding/json"

	"github.com/yestefir/micup.go/src/model"
)

type GuildCreate struct {
	model.Guild
	Unavailable       bool               `json:"unavailable,omitempty"`
	JoinedAt          string             `json:"joined_at,omitempty"`
	Channels          []model.Channel    `json:"channels,omitempty"`
	Threads           []model.Channel    `json:"threads,omitempty"`
	Members           []model.Member     `json:"members,omitempty"`
	VoiceStates       []model.VoiceState `json:"voice_states,omitempty"`
	Presences         []PresenceUpdate   `json:"presences,omitempty"`
	Properties        *model.Guild       `json:"properties,omitempty"`
	Visiting          bool               `json:"visiting,omitempty"`
	UnavailableHidden bool               `json:"unavailable_hidden,omitempty"`
}

func (g *GuildCreate) UnmarshalJSON(data []byte) error {
	type plain GuildCreate
	err := json.Unmarshal(data, (*plain)(g))
	g.Guild = withProperties(g.Guild, g.Properties)
	return err
}

type GuildUpdate struct {
	model.Guild
	GuildID    string       `json:"guild_id,omitempty"`
	Properties *model.Guild `json:"properties,omitempty"`
}

func (g *GuildUpdate) UnmarshalJSON(data []byte) error {
	type plain GuildUpdate
	err := json.Unmarshal(data, (*plain)(g))
	if g.ID == "" {
		g.ID = g.GuildID
	}
	g.Guild = withProperties(g.Guild, g.Properties)
	return err
}

type GuildDelete struct {
	ID                string `json:"id"`
	Unavailable       bool   `json:"unavailable,omitempty"`
	GuildID           string `json:"guild_id,omitempty"`
	UnavailableHidden bool   `json:"unavailable_hidden,omitempty"`
}

type GuildEmojisUpdate struct {
	GuildID string        `json:"guild_id"`
	Emojis  []model.Emoji `json:"emojis"`
}

type GuildStickersUpdate struct {
	GuildID  string          `json:"guild_id"`
	Stickers []model.Sticker `json:"stickers"`
}

type GuildCount struct {
	GuildID     string `json:"guild_id"`
	MemberCount int    `json:"member_count"`
	OnlineCount int    `json:"online_count"`
}

type GuildCountsUpdate struct {
	Counts []GuildCount `json:"counts"`
	Nonce  string       `json:"nonce,omitempty"`
}

type ChannelMemberCount struct {
	GuildID     string `json:"guild_id"`
	ChannelID   string `json:"channel_id"`
	MemberCount int    `json:"member_count"`
	OnlineCount int    `json:"online_count"`
}

type ChannelMemberCountsUpdate struct {
	Counts []ChannelMemberCount `json:"counts"`
	Nonce  string               `json:"nonce,omitempty"`
}

type PassiveUpdates struct {
	GuildID     string             `json:"guild_id"`
	Channels    map[string]string  `json:"channels,omitempty"`
	VoiceStates []model.VoiceState `json:"voice_states,omitempty"`
}

func withProperties(guild model.Guild, properties *model.Guild) model.Guild {
	if properties == nil {
		return guild
	}
	merged := *properties
	merged.ID = guild.ID
	if merged.Name == "" {
		merged.Name = guild.Name
	}
	if merged.OwnerID == "" {
		merged.OwnerID = guild.OwnerID
	}
	merged.Roles, merged.Emojis, merged.Stickers = guild.Roles, guild.Emojis, guild.Stickers
	merged.MemberCount = max(merged.MemberCount, guild.MemberCount)
	merged.OnlineCount = max(merged.OnlineCount, guild.OnlineCount)
	return merged
}
