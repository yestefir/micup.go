package model

import "encoding/json"

const (
	AuditGuildUpdate            = 1
	AuditChannelCreate          = 10
	AuditChannelUpdate          = 11
	AuditChannelDelete          = 12
	AuditChannelOverwriteCreate = 13
	AuditChannelOverwriteUpdate = 14
	AuditChannelOverwriteDelete = 15
	AuditMemberKick             = 20
	AuditMemberPrune            = 21
	AuditMemberBanAdd           = 22
	AuditMemberBanRemove        = 23
	AuditMemberUpdate           = 24
	AuditMemberRoleUpdate       = 25
	AuditMemberMove             = 26
	AuditMemberDisconnect       = 27
	AuditBotAdd                 = 28
	AuditRoleCreate             = 30
	AuditRoleUpdate             = 31
	AuditRoleDelete             = 32
	AuditInviteCreate           = 40
	AuditInviteUpdate           = 41
	AuditInviteDelete           = 42
	AuditWebhookCreate          = 50
	AuditWebhookUpdate          = 51
	AuditWebhookDelete          = 52
	AuditEmojiCreate            = 60
	AuditEmojiUpdate            = 61
	AuditEmojiDelete            = 62
	AuditMessageDelete          = 72
	AuditMessageBulkDelete      = 73
	AuditMessagePin             = 74
	AuditMessageUnpin           = 75
	AuditStickerCreate          = 90
	AuditStickerUpdate          = 91
	AuditStickerDelete          = 92
	AuditThreadCreate           = 110
	AuditThreadUpdate           = 111
	AuditThreadDelete           = 112

	MaxAuditLogEntries = 100
)

type AuditLogChange struct {
	Key      string          `json:"key"`
	OldValue json.RawMessage `json:"old_value,omitempty"`
	NewValue json.RawMessage `json:"new_value,omitempty"`
}

type AuditLogOptions struct {
	ChannelID            string      `json:"channel_id,omitempty"`
	MessageID            string      `json:"message_id,omitempty"`
	ID                   string      `json:"id,omitempty"`
	Type                 json.Number `json:"type,omitempty"`
	Count                json.Number `json:"count,omitempty"`
	DeleteMessageSeconds json.Number `json:"delete_message_seconds,omitempty"`
	InviterID            string      `json:"inviter_id,omitempty"`
	MaxAge               json.Number `json:"max_age,omitempty"`
	MaxUses              json.Number `json:"max_uses,omitempty"`
	RoleName             string      `json:"role_name,omitempty"`
	Temporary            bool        `json:"temporary,omitempty"`
}

type AuditLogEntry struct {
	ID         string           `json:"id"`
	ActionType int              `json:"action_type"`
	UserID     string           `json:"user_id,omitempty"`
	TargetID   string           `json:"target_id,omitempty"`
	Reason     string           `json:"reason,omitempty"`
	Options    *AuditLogOptions `json:"options,omitempty"`
	Changes    []AuditLogChange `json:"changes,omitempty"`
}

type AuditLogWebhook struct {
	ID         string  `json:"id"`
	Type       int     `json:"type"`
	GuildID    string  `json:"guild_id,omitempty"`
	ChannelID  string  `json:"channel_id,omitempty"`
	Name       string  `json:"name"`
	AvatarHash *string `json:"avatar_hash,omitempty"`
}

type AuditLog struct {
	Entries  []AuditLogEntry   `json:"audit_log_entries"`
	Users    []User            `json:"users,omitempty"`
	Webhooks []AuditLogWebhook `json:"webhooks,omitempty"`
	Threads  []Channel         `json:"threads,omitempty"`
}

type AuditLogQuery struct {
	Limit      int
	Before     string
	After      string
	UserID     string
	ActionType int
}
