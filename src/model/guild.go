package model

type Guild struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name"`
	OwnerID                     string   `json:"owner_id,omitempty"`
	Icon                        *string  `json:"icon,omitempty"`
	Banner                      *string  `json:"banner,omitempty"`
	Splash                      *string  `json:"splash,omitempty"`
	Features                    []string `json:"features,omitempty"`
	AFKChannelID                *string  `json:"afk_channel_id,omitempty"`
	AFKTimeout                  int      `json:"afk_timeout,omitempty"`
	SystemChannelID             *string  `json:"system_channel_id,omitempty"`
	SystemChannelFlags          int      `json:"system_channel_flags,omitempty"`
	RulesChannelID              *string  `json:"rules_channel_id,omitempty"`
	VanityURLCode               *string  `json:"vanity_url_code,omitempty"`
	VerificationLevel           int      `json:"verification_level,omitempty"`
	MFALevel                    int      `json:"mfa_level,omitempty"`
	NSFW                        bool     `json:"nsfw,omitempty"`
	NSFWLevel                   int      `json:"nsfw_level,omitempty"`
	ExplicitContentFilter       int      `json:"explicit_content_filter,omitempty"`
	DefaultMessageNotifications int      `json:"default_message_notifications,omitempty"`
	PremiumTier                 int      `json:"premium_tier,omitempty"`
	PremiumSubscriptionCount    int      `json:"premium_subscription_count,omitempty"`
	MemberCount                 int      `json:"member_count,omitempty"`
	OnlineCount                 int      `json:"online_count,omitempty"`
	Roles                       []Role   `json:"roles,omitempty"`
	Emojis                      []Emoji  `json:"emojis,omitempty"`
}

type Invite struct {
	Code          string   `json:"code"`
	Type          int      `json:"type"`
	GuildID       string   `json:"guild_id,omitempty"`
	ChannelID     string   `json:"channel_id,omitempty"`
	Guild         *Guild   `json:"guild,omitempty"`
	Channel       *Channel `json:"channel,omitempty"`
	Inviter       *User    `json:"inviter,omitempty"`
	Uses          int      `json:"uses"`
	MaxUses       int      `json:"max_uses"`
	MaxAge        int      `json:"max_age"`
	Temporary     bool     `json:"temporary"`
	CreatedAt     string   `json:"created_at,omitempty"`
	ExpiresAt     *string  `json:"expires_at,omitempty"`
	MemberCount   int      `json:"member_count,omitempty"`
	PresenceCount int      `json:"presence_count,omitempty"`
}
