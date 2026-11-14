package events

import "github.com/yestefir/micup.go/src/model"

type PresenceUpdate struct {
	User         model.User          `json:"user"`
	GuildID      string              `json:"guild_id,omitempty"`
	Status       string              `json:"status"`
	AFK          bool                `json:"afk"`
	Mobile       bool                `json:"mobile"`
	CustomStatus *model.CustomStatus `json:"custom_status,omitempty"`
}

type TypingStart struct {
	ChannelID string        `json:"channel_id"`
	UserID    string        `json:"user_id"`
	Timestamp int64         `json:"timestamp"`
	GuildID   string        `json:"guild_id,omitempty"`
	Member    *model.Member `json:"member,omitempty"`
}

type PresenceUpdateBulk struct {
	GuildID   string           `json:"guild_id"`
	Presences []PresenceUpdate `json:"presences"`
}

const (
	RelationshipFriend          = 1
	RelationshipBlocked         = 2
	RelationshipIncomingRequest = 3
	RelationshipOutgoingRequest = 4
)

type Relationship struct {
	ID       string      `json:"id"`
	Type     int         `json:"type"`
	User     *model.User `json:"user,omitempty"`
	Since    string      `json:"since,omitempty"`
	Nickname *string     `json:"nickname,omitempty"`
}

type RelationshipRemove struct {
	ID string `json:"id"`
}

type Call struct {
	ChannelID   string             `json:"channel_id"`
	MessageID   string             `json:"message_id,omitempty"`
	Region      *string            `json:"region,omitempty"`
	Ringing     []string           `json:"ringing,omitempty"`
	VoiceStates []model.VoiceState `json:"voice_states,omitempty"`
	Recipients  []string           `json:"recipients,omitempty"`
	CreatedAt   int64              `json:"created_at,omitempty"`
}

type CallDelete struct {
	ChannelID   string `json:"channel_id"`
	Unavailable bool   `json:"unavailable,omitempty"`
}

type EntranceSoundPlay struct {
	UserID      string `json:"user_id"`
	ChannelID   string `json:"channel_id"`
	GuildID     string `json:"guild_id,omitempty"`
	SoundID     string `json:"sound_id"`
	Hash        string `json:"hash,omitempty"`
	URL         string `json:"url"`
	DurationMS  int    `json:"duration_ms,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}
