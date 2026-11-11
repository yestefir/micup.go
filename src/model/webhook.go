package model

type Webhook struct {
	ID            string         `json:"id"`
	GuildID       string         `json:"guild_id"`
	ChannelID     string         `json:"channel_id"`
	Name          string         `json:"name"`
	Avatar        *string        `json:"avatar,omitempty"`
	Type          int            `json:"type"`
	Token         string         `json:"token,omitempty"`
	User          *User          `json:"user,omitempty"`
	SourceGuild   *WebhookSource `json:"source_guild,omitempty"`
	SourceChannel *WebhookSource `json:"source_channel,omitempty"`
}

const (
	WebhookTypeIncoming        = 1
	WebhookTypeChannelFollower = 2
	MaxWebhookName             = 80
)

type WebhookSource struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon,omitempty"`
}

type WebhookEdit struct {
	Name      Optional[string]
	Avatar    Optional[string]
	ChannelID Optional[string]
}

type WebhookMessage struct {
	MessageOptions
	Username    string
	AvatarURL   string
	ThreadID    string
	Wait        bool
	ThreadName  string
	AppliedTags []string
}
