package events

import "github.com/yestefir/micup.go/src/model"

type MessageCreate struct {
	model.Message
	ChannelType int               `json:"channel_type"`
	MentionHere bool              `json:"mention_here,omitempty"`
	Member      *model.Member     `json:"member,omitempty"`
	Nicks       map[string]string `json:"nicks,omitempty"`
}

type MessageUpdate = MessageCreate

type MessageDelete struct {
	ID        string        `json:"id"`
	ChannelID string        `json:"channel_id"`
	GuildID   string        `json:"guild_id,omitempty"`
	Content   *string       `json:"content,omitempty"`
	AuthorID  string        `json:"author_id,omitempty"`
	Member    *model.Member `json:"member,omitempty"`
}

type MessageDeleteBulk struct {
	IDs       []string `json:"ids"`
	ChannelID string   `json:"channel_id"`
	GuildID   string   `json:"guild_id,omitempty"`
}
