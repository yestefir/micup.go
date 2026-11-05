package model

const (
	ChannelTypeGuildText          = 0
	ChannelTypeDM                 = 1
	ChannelTypeGuildVoice         = 2
	ChannelTypeGroupDM            = 3
	ChannelTypeGuildCategory      = 4
	ChannelTypeGuildAnnouncement  = 5
	ChannelTypeAnnouncementThread = 10
	ChannelTypePublicThread       = 11
	ChannelTypePrivateThread      = 12
	ChannelTypeGuildForum         = 15
	ChannelTypeGuildMedia         = 16
	ChannelTypeGuildLink          = 998
	ChannelTypePersonalNotes      = 999
)

type Channel struct {
	ID                   string                `json:"id"`
	GuildID              string                `json:"guild_id"`
	Name                 string                `json:"name"`
	Type                 int                   `json:"type"`
	Topic                *string               `json:"topic,omitempty"`
	Position             int                   `json:"position,omitempty"`
	ParentID             string                `json:"parent_id,omitempty"`
	Bitrate              int                   `json:"bitrate,omitempty"`
	UserLimit            int                   `json:"user_limit,omitempty"`
	RTCRegion            *string               `json:"rtc_region,omitempty"`
	NSFW                 bool                  `json:"nsfw,omitempty"`
	RateLimitPerUser     int                   `json:"rate_limit_per_user,omitempty"`
	LastMessageID        string                `json:"last_message_id,omitempty"`
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites,omitempty"`
	Recipients           []User                `json:"recipients,omitempty"`
}
