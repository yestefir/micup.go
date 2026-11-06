package events

import "github.com/yestefir/micup.go/src/model"

type PresenceUpdate struct {
	User         model.User          `json:"user"`
	GuildID      string              `json:"guild_id,omitempty"`
	Status       string              `json:"status"`
	AFK          bool                `json:"afk"`
	Mobile       bool                `json:"mobile"`
	CustomStatus *model.CustomStatus `json:"custom_status,omitempty"`
}

type TypingStart struct {
	ChannelID string        `json:"channel_id"`
	UserID    string        `json:"user_id"`
	Timestamp int64         `json:"timestamp"`
	GuildID   string        `json:"guild_id,omitempty"`
	Member    *model.Member `json:"member,omitempty"`
}
