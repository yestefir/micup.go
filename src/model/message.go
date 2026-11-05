package model

const (
	MessageTypeDefault = 0
	MessageTypeReply   = 19
)

const (
	MessageFlagEphemeral = 64
	MaxEmbeds            = 10
)

type MessageCreate struct {
	ID        string `json:"id"`
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
	Type      int    `json:"type"`
	Author    User   `json:"author"`
	Mentions  []User `json:"mentions"`
}

type Message struct {
	ID        string  `json:"id"`
	ChannelID string  `json:"channel_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds,omitempty"`
	Flags     int     `json:"flags,omitempty"`
}

type MessageOptions struct {
	Content      string  `json:"content,omitempty"`
	Embeds       []Embed `json:"embeds,omitempty"`
	ReplyTo      string  `json:"-"`
	Nonce        string  `json:"nonce,omitempty"`
	EnforceNonce bool    `json:"enforce_nonce,omitempty"`
}
