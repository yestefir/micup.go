package events

import "github.com/yestefir/micup.go/src/model"

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

type (
	AuditLogChange  = model.AuditLogChange
	AuditLogOptions = model.AuditLogOptions
)
