package micup

import "encoding/json"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Channel struct {
	ID      string `json:"id"`
	GuildID string `json:"guild_id"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
}

type MessageCreate struct {
	ID        string          `json:"id"`
	GuildID   string          `json:"guild_id"`
	ChannelID string          `json:"channel_id"`
	Content   string          `json:"content"`
	Author    User            `json:"author"`
	Mentions  []User          `json:"mentions"`
	Raw       json.RawMessage `json:"-"`
}

type Embed struct {
	Title       string       `json:"title,omitempty"`
	Type        string       `json:"type,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
	Color       int          `json:"color,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Image       *EmbedMedia  `json:"image,omitempty"`
	Thumbnail   *EmbedMedia  `json:"thumbnail,omitempty"`
	Author      *EmbedAuthor `json:"author,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
}

type EmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedMedia struct {
	URL string `json:"url"`
}

type EmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type MessageOptions struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
	ReplyTo string  `json:"-"`
}

type Interaction struct {
	ID            string          `json:"id"`
	ApplicationID string          `json:"application_id"`
	Type          int             `json:"type"`
	Data          json.RawMessage `json:"data,omitempty"`
	GuildID       string          `json:"guild_id,omitempty"`
	ChannelID     string          `json:"channel_id,omitempty"`
	Token         string          `json:"token"`
	Version       int             `json:"version,omitempty"`
	Member        json.RawMessage `json:"member,omitempty"`
	User          json.RawMessage `json:"user,omitempty"`
}

type VoiceState struct {
	GuildID   string `json:"guild_id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
}

type VoiceServerUpdate struct {
	GuildID      string `json:"guild_id"`
	Endpoint     string `json:"endpoint"`
	Token        string `json:"token"`
	ConnectionID string `json:"connection_id"`
}
