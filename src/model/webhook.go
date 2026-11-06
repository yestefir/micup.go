package model

type Webhook struct {
	ID        string  `json:"id"`
	GuildID   string  `json:"guild_id"`
	ChannelID string  `json:"channel_id"`
	Name      string  `json:"name"`
	Avatar    *string `json:"avatar,omitempty"`
	Type      int     `json:"type"`
	Token     string  `json:"token,omitempty"`
	User      *User   `json:"user,omitempty"`
}

type WebhookMessage struct {
	MessageOptions
	Username  string
	AvatarURL string
	ThreadID  string
	Wait      bool
}
