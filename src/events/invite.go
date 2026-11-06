package events

type InviteDelete struct {
	Code      string `json:"code"`
	GuildID   string `json:"guild_id,omitempty"`
	ChannelID string `json:"channel_id"`
}

type WebhooksUpdate struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
}
