package micup

import (
	"context"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
)

const (
	EventReady                     = gateway.EventReady
	EventResumed                   = gateway.EventResumed
	EventRateLimited               = gateway.EventRateLimited
	EventSessionsReplace           = gateway.EventSessionsReplace
	EventGatewayRolloutUpdate      = gateway.EventGatewayRolloutUpdate
	EventGuildCreate               = gateway.EventGuildCreate
	EventGuildSync                 = gateway.EventGuildSync
	EventGuildUpdate               = gateway.EventGuildUpdate
	EventGuildDelete               = gateway.EventGuildDelete
	EventGuildAuditLogEntryCreate  = gateway.EventGuildAuditLogEntryCreate
	EventGuildRoleCreate           = gateway.EventGuildRoleCreate
	EventGuildRoleUpdate           = gateway.EventGuildRoleUpdate
	EventGuildRoleUpdateBulk       = gateway.EventGuildRoleUpdateBulk
	EventGuildRoleDelete           = gateway.EventGuildRoleDelete
	EventGuildMemberAdd            = gateway.EventGuildMemberAdd
	EventGuildMemberUpdate         = gateway.EventGuildMemberUpdate
	EventGuildMemberRemove         = gateway.EventGuildMemberRemove
	EventGuildMembersChunk         = gateway.EventGuildMembersChunk
	EventGuildBanAdd               = gateway.EventGuildBanAdd
	EventGuildBanRemove            = gateway.EventGuildBanRemove
	EventChannelCreate             = gateway.EventChannelCreate
	EventChannelUpdate             = gateway.EventChannelUpdate
	EventChannelUpdateBulk         = gateway.EventChannelUpdateBulk
	EventChannelDelete             = gateway.EventChannelDelete
	EventChannelPinsUpdate         = gateway.EventChannelPinsUpdate
	EventMessageCreate             = gateway.EventMessageCreate
	EventMessageUpdate             = gateway.EventMessageUpdate
	EventMessageDelete             = gateway.EventMessageDelete
	EventMessageDeleteBulk         = gateway.EventMessageDeleteBulk
	EventReactionAdd               = gateway.EventReactionAdd
	EventReactionRemove            = gateway.EventReactionRemove
	EventReactionRemoveAll         = gateway.EventReactionRemoveAll
	EventReactionRemoveEmoji       = gateway.EventReactionRemoveEmoji
	EventUserUpdate                = gateway.EventUserUpdate
	EventPresenceUpdate            = gateway.EventPresenceUpdate
	EventTypingStart               = gateway.EventTypingStart
	EventInviteCreate              = gateway.EventInviteCreate
	EventInviteDelete              = gateway.EventInviteDelete
	EventWebhooksUpdate            = gateway.EventWebhooksUpdate
	EventInteractionCreate         = gateway.EventInteractionCreate
	EventApplicationCommandsUpdate = gateway.EventApplicationCommandsUpdate
	EventVoiceStateUpdate          = gateway.EventVoiceStateUpdate
	EventVoiceServerUpdate         = gateway.EventVoiceServerUpdate
)

type (
	Ready                     = events.Ready
	ReadyApplication          = events.ReadyApplication
	Resumed                   = events.Resumed
	ClientSession             = events.ClientSession
	SessionsReplace           = events.SessionsReplace
	GatewayRollout            = events.GatewayRollout
	RateLimited               = events.RateLimited
	RateLimitedMeta           = events.RateLimitedMeta
	GuildCreate               = events.GuildCreate
	GuildUpdate               = events.GuildUpdate
	GuildDelete               = events.GuildDelete
	AuditLogEntryCreate       = events.AuditLogEntryCreate
	AuditLogChange            = events.AuditLogChange
	AuditLogOptions           = events.AuditLogOptions
	RoleEvent                 = events.RoleEvent
	RoleUpdateBulk            = events.RoleUpdateBulk
	RoleDelete                = events.RoleDelete
	GuildMember               = events.GuildMember
	GuildMemberRemove         = events.GuildMemberRemove
	MembersChunk              = events.MembersChunk
	GuildBan                  = events.GuildBan
	ChannelUpdateBulk         = events.ChannelUpdateBulk
	ChannelPinsUpdate         = events.ChannelPinsUpdate
	MessageCreate             = events.MessageCreate
	MessageUpdate             = events.MessageUpdate
	MessageDelete             = events.MessageDelete
	MessageDeleteBulk         = events.MessageDeleteBulk
	ReactionEvent             = events.ReactionEvent
	ReactionRemoveAll         = events.ReactionRemoveAll
	ReactionRemoveEmoji       = events.ReactionRemoveEmoji
	PresenceUpdate            = events.PresenceUpdate
	TypingStart               = events.TypingStart
	InviteDelete              = events.InviteDelete
	WebhooksUpdate            = events.WebhooksUpdate
	ApplicationCommandsUpdate = events.ApplicationCommandsUpdate
	EventMux                  = events.Mux
	EventListener             = events.Listener
)

func KnownEvents() []string { return gateway.KnownEvents() }

func DecodeEvent(event Event) (any, error) { return events.Decode(event) }

func EventAs[T any](event Event) (T, error) { return events.As[T](event) }

func NewEventMux() *EventMux { return events.NewMux() }

func Listen[T any](fn func(context.Context, T) error) EventListener { return events.Listen(fn) }

const (
	EventGuildEmojisUpdate         = gateway.EventGuildEmojisUpdate
	EventGuildStickersUpdate       = gateway.EventGuildStickersUpdate
	EventGuildMemberListUpdate     = gateway.EventGuildMemberListUpdate
	EventGuildCountsUpdate         = gateway.EventGuildCountsUpdate
	EventChannelRecipientAdd       = gateway.EventChannelRecipientAdd
	EventChannelRecipientRemove    = gateway.EventChannelRecipientRemove
	EventChannelMemberCountsUpdate = gateway.EventChannelMemberCountsUpdate
	EventPresenceUpdateBulk        = gateway.EventPresenceUpdateBulk
	EventPassiveUpdates            = gateway.EventPassiveUpdates
	EventReactionAddMany           = gateway.EventReactionAddMany
	EventThreadCreate              = gateway.EventThreadCreate
	EventThreadUpdate              = gateway.EventThreadUpdate
	EventThreadDelete              = gateway.EventThreadDelete
	EventThreadListSync            = gateway.EventThreadListSync
	EventThreadMemberUpdate        = gateway.EventThreadMemberUpdate
	EventThreadMembersUpdate       = gateway.EventThreadMembersUpdate
	EventThreadMemberListUpdate    = gateway.EventThreadMemberListUpdate
	EventCallCreate                = gateway.EventCallCreate
	EventCallUpdate                = gateway.EventCallUpdate
	EventCallDelete                = gateway.EventCallDelete
	EventEntranceSoundPlay         = gateway.EventEntranceSoundPlay
	EventRelationshipAdd           = gateway.EventRelationshipAdd
	EventRelationshipUpdate        = gateway.EventRelationshipUpdate
	EventRelationshipRemove        = gateway.EventRelationshipRemove
)

type (
	GuildEmojisUpdate         = events.GuildEmojisUpdate
	GuildStickersUpdate       = events.GuildStickersUpdate
	GuildCount                = events.GuildCount
	GuildCountsUpdate         = events.GuildCountsUpdate
	ChannelMemberCount        = events.ChannelMemberCount
	ChannelMemberCountsUpdate = events.ChannelMemberCountsUpdate
	PassiveUpdates            = events.PassiveUpdates
	ReactionAddition          = events.ReactionAddition
	ReactionAddMany           = events.ReactionAddMany
	ChannelRecipient          = events.ChannelRecipient
	ThreadCreate              = events.ThreadCreate
	ThreadDelete              = events.ThreadDelete
	ThreadListSync            = events.ThreadListSync
	ThreadMemberPresence      = events.ThreadMemberPresence
	ThreadMembersUpdate       = events.ThreadMembersUpdate
	ThreadMemberListUpdate    = events.ThreadMemberListUpdate
	MemberListGroup           = events.MemberListGroup
	MemberListMember          = events.MemberListMember
	MemberListItem            = events.MemberListItem
	MemberListOperation       = events.MemberListOperation
	MemberListUpdate          = events.MemberListUpdate
	PresenceUpdateBulk        = events.PresenceUpdateBulk
	Relationship              = events.Relationship
	RelationshipRemove        = events.RelationshipRemove
	Call                      = events.Call
	CallDelete                = events.CallDelete
	EntranceSoundPlay         = events.EntranceSoundPlay
	Subscription              = gateway.Subscription
)

const (
	RelationshipFriend          = events.RelationshipFriend
	RelationshipBlocked         = events.RelationshipBlocked
	RelationshipIncomingRequest = events.RelationshipIncomingRequest
	RelationshipOutgoingRequest = events.RelationshipOutgoingRequest
)
