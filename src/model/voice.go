package model

type VoiceState struct {
	GuildID   string `json:"guild_id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
}

type VoiceServerUpdate struct {
	GuildID      string `json:"guild_id"`
	ChannelID    string `json:"channel_id,omitempty"`
	Endpoint     string `json:"endpoint"`
	Token        string `json:"token"`
	ConnectionID string `json:"connection_id"`
}
