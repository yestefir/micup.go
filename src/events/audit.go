package events

import "encoding/json"

type AuditLogEntryCreate struct {
	ID         string           `json:"id"`
	GuildID    string           `json:"guild_id"`
	UserID     string           `json:"user_id,omitempty"`
	TargetID   string           `json:"target_id,omitempty"`
	ActionType int              `json:"action_type"`
	Reason     string           `json:"reason,omitempty"`
	Changes    []AuditLogChange `json:"changes,omitempty"`
	Options    *AuditLogOptions `json:"options,omitempty"`
}

type AuditLogChange struct {
	Key      string          `json:"key"`
	OldValue json.RawMessage `json:"old_value,omitempty"`
	NewValue json.RawMessage `json:"new_value,omitempty"`
}

type AuditLogOptions struct {
	ChannelID string      `json:"channel_id,omitempty"`
	MessageID string      `json:"message_id,omitempty"`
	ID        string      `json:"id,omitempty"`
	Type      json.Number `json:"type,omitempty"`
	Count     json.Number `json:"count,omitempty"`
}
