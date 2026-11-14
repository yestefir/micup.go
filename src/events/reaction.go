package events

import "github.com/yestefir/micup.go/src/model"

type ReactionEvent struct {
	UserID    string        `json:"user_id"`
	ChannelID string        `json:"channel_id"`
	MessageID string        `json:"message_id"`
	GuildID   string        `json:"guild_id,omitempty"`
	Emoji     model.Emoji   `json:"emoji"`
	Member    *model.Member `json:"member,omitempty"`
	Burst     bool          `json:"burst,omitempty"`
	Type      int           `json:"type,omitempty"`
	SessionID string        `json:"session_id,omitempty"`
}

type ReactionAddition struct {
	UserID string        `json:"user_id"`
	Emoji  model.Emoji   `json:"emoji"`
	Member *model.Member `json:"member,omitempty"`
}

type ReactionAddMany struct {
	ChannelID string             `json:"channel_id"`
	MessageID string             `json:"message_id"`
	GuildID   string             `json:"guild_id,omitempty"`
	Reactions []ReactionAddition `json:"reactions"`
}

type ReactionRemoveAll struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	GuildID   string `json:"guild_id,omitempty"`
}

type ReactionRemoveEmoji struct {
	ChannelID string      `json:"channel_id"`
	MessageID string      `json:"message_id"`
	GuildID   string      `json:"guild_id,omitempty"`
	Emoji     model.Emoji `json:"emoji"`
}
