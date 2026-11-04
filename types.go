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
