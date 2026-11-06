package micup

import (
	"context"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
)

const (
	EventReady                    = gateway.EventReady
	EventResumed                  = gateway.EventResumed
	EventRateLimited              = gateway.EventRateLimited
	EventSessionsReplace          = gateway.EventSessionsReplace
	EventGuildCreate              = gateway.EventGuildCreate
	EventGuildSync                = gateway.EventGuildSync
	EventGuildUpdate              = gateway.EventGuildUpdate
	EventGuildDelete              = gateway.EventGuildDelete
	EventGuildAuditLogEntryCreate = gateway.EventGuildAuditLogEntryCreate
	EventGuildRoleCreate          = gateway.EventGuildRoleCreate
	EventGuildRoleUpdate          = gateway.EventGuildRoleUpdate
	EventGuildRoleUpdateBulk      = gateway.EventGuildRoleUpdateBulk
	EventGuildRoleDelete          = gateway.EventGuildRoleDelete
	EventGuildMemberAdd           = gateway.EventGuildMemberAdd
	EventGuildMemberUpdate        = gateway.EventGuildMemberUpdate
	EventGuildMemberRemove        = gateway.EventGuildMemberRemove
	EventGuildMembersChunk        = gateway.EventGuildMembersChunk
	EventGuildBanAdd              = gateway.EventGuildBanAdd
	EventGuildBanRemove           = gateway.EventGuildBanRemove
	EventChannelCreate            = gateway.EventChannelCreate
	EventChannelUpdate            = gateway.EventChannelUpdate
	EventChannelUpdateBulk        = gateway.EventChannelUpdateBulk
	EventChannelDelete            = gateway.EventChannelDelete
	EventChannelPinsUpdate        = gateway.EventChannelPinsUpdate
	EventMessageCreate            = gateway.EventMessageCreate
	EventMessageUpdate            = gateway.EventMessageUpdate
	EventMessageDelete            = gateway.EventMessageDelete
	EventMessageDeleteBulk        = gateway.EventMessageDeleteBulk
	EventReactionAdd              = gateway.EventReactionAdd
	EventReactionRemove           = gateway.EventReactionRemove
	EventReactionRemoveAll        = gateway.EventReactionRemoveAll
	EventReactionRemoveEmoji      = gateway.EventReactionRemoveEmoji
	EventUserUpdate               = gateway.EventUserUpdate
	EventPresenceUpdate           = gateway.EventPresenceUpdate
	EventTypingStart              = gateway.EventTypingStart
	EventInviteCreate             = gateway.EventInviteCreate
	EventInviteDelete             = gateway.EventInviteDelete
	EventWebhooksUpdate           = gateway.EventWebhooksUpdate
	EventInteractionCreate        = gateway.EventInteractionCreate
	EventVoiceStateUpdate         = gateway.EventVoiceStateUpdate
	EventVoiceServerUpdate        = gateway.EventVoiceServerUpdate
)

type (
	Ready               = events.Ready
	ReadyApplication    = events.ReadyApplication
	Resumed             = events.Resumed
	ClientSession       = events.ClientSession
	SessionsReplace     = events.SessionsReplace
	RateLimited         = events.RateLimited
	RateLimitedMeta     = events.RateLimitedMeta
	GuildCreate         = events.GuildCreate
	GuildUpdate         = events.GuildUpdate
	GuildDelete         = events.GuildDelete
	AuditLogEntryCreate = events.AuditLogEntryCreate
	AuditLogChange      = events.AuditLogChange
	AuditLogOptions     = events.AuditLogOptions
	RoleEvent           = events.RoleEvent
	RoleUpdateBulk      = events.RoleUpdateBulk
	RoleDelete          = events.RoleDelete
	GuildMember         = events.GuildMember
	GuildMemberRemove   = events.GuildMemberRemove
	MembersChunk        = events.MembersChunk
	GuildBan            = events.GuildBan
	ChannelUpdateBulk   = events.ChannelUpdateBulk
	ChannelPinsUpdate   = events.ChannelPinsUpdate
	MessageCreate       = events.MessageCreate
	MessageUpdate       = events.MessageUpdate
	MessageDelete       = events.MessageDelete
	MessageDeleteBulk   = events.MessageDeleteBulk
	ReactionEvent       = events.ReactionEvent
	ReactionRemoveAll   = events.ReactionRemoveAll
	ReactionRemoveEmoji = events.ReactionRemoveEmoji
	PresenceUpdate      = events.PresenceUpdate
	TypingStart         = events.TypingStart
	InviteDelete        = events.InviteDelete
	WebhooksUpdate      = events.WebhooksUpdate
	EventMux            = events.Mux
	EventListener       = events.Listener
)

func KnownEvents() []string { return gateway.KnownEvents() }

func DecodeEvent(event Event) (any, error) { return events.Decode(event) }

func EventAs[T any](event Event) (T, error) { return events.As[T](event) }

func NewEventMux() *EventMux { return events.NewMux() }

func Listen[T any](fn func(context.Context, T) error) EventListener { return events.Listen(fn) }
