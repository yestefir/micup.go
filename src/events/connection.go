package events

import "github.com/yestefir/micup.go/src/model"

type Ready struct {
	Version          int              `json:"v"`
	User             model.User       `json:"user"`
	Guilds           []GuildCreate    `json:"guilds"`
	SessionID        string           `json:"session_id"`
	ResumeGatewayURL string           `json:"resume_gateway_url,omitempty"`
	Shard            []int            `json:"shard,omitempty"`
	Application      ReadyApplication `json:"application"`
	Sessions         []ClientSession  `json:"sessions,omitempty"`
	Presences        []PresenceUpdate `json:"presences,omitempty"`
	PrivateChannels  []model.Channel  `json:"private_channels,omitempty"`
}

type ReadyApplication struct {
	ID    string `json:"id"`
	Flags uint64 `json:"flags"`
}

type Resumed struct{}

type ClientSession struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	AFK       bool   `json:"afk"`
	Mobile    bool   `json:"mobile"`
}

type SessionsReplace []ClientSession

type RateLimited struct {
	Opcode     int             `json:"opcode"`
	RetryAfter float64         `json:"retry_after"`
	Meta       RateLimitedMeta `json:"meta"`
}

type RateLimitedMeta struct {
	GuildID string `json:"guild_id,omitempty"`
	Nonce   string `json:"nonce,omitempty"`
}
