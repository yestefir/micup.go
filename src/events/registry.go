package events

import (
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

var registry = map[string]payload{
	gateway.EventReady:           of[Ready](),
	gateway.EventResumed:         of[Resumed](),
	gateway.EventRateLimited:     of[RateLimited](),
	gateway.EventSessionsReplace: of[SessionsReplace](),

	gateway.EventGuildCreate:              of[GuildCreate](),
	gateway.EventGuildSync:                of[GuildCreate](),
	gateway.EventGuildUpdate:              of[GuildUpdate](),
	gateway.EventGuildDelete:              of[GuildDelete](),
	gateway.EventGuildAuditLogEntryCreate: of[AuditLogEntryCreate](),

	gateway.EventGuildRoleCreate:     of[RoleEvent](),
	gateway.EventGuildRoleUpdate:     of[RoleEvent](),
	gateway.EventGuildRoleUpdateBulk: of[RoleUpdateBulk](),
	gateway.EventGuildRoleDelete:     of[RoleDelete](),

	gateway.EventGuildMemberAdd:    of[GuildMember](),
	gateway.EventGuildMemberUpdate: of[GuildMember](),
	gateway.EventGuildMemberRemove: of[GuildMemberRemove](),
	gateway.EventGuildMembersChunk: of[MembersChunk](),
	gateway.EventGuildBanAdd:       of[GuildBan](),
	gateway.EventGuildBanRemove:    of[GuildBan](),

	gateway.EventChannelCreate:     of[model.Channel](),
	gateway.EventChannelUpdate:     of[model.Channel](),
	gateway.EventChannelUpdateBulk: of[ChannelUpdateBulk](),
	gateway.EventChannelDelete:     of[model.Channel](),
	gateway.EventChannelPinsUpdate: of[ChannelPinsUpdate](),

	gateway.EventMessageCreate:     of[MessageCreate](),
	gateway.EventMessageUpdate:     of[MessageUpdate](),
	gateway.EventMessageDelete:     of[MessageDelete](),
	gateway.EventMessageDeleteBulk: of[MessageDeleteBulk](),

	gateway.EventReactionAdd:         of[ReactionEvent](),
	gateway.EventReactionRemove:      of[ReactionEvent](),
	gateway.EventReactionRemoveAll:   of[ReactionRemoveAll](),
	gateway.EventReactionRemoveEmoji: of[ReactionRemoveEmoji](),

	gateway.EventUserUpdate:     of[model.User](),
	gateway.EventPresenceUpdate: of[PresenceUpdate](),
	gateway.EventTypingStart:    of[TypingStart](),

	gateway.EventInviteCreate:   of[model.Invite](),
	gateway.EventInviteDelete:   of[InviteDelete](),
	gateway.EventWebhooksUpdate: of[WebhooksUpdate](),

	gateway.EventInteractionCreate:         of[model.Interaction](),
	gateway.EventApplicationCommandsUpdate: of[ApplicationCommandsUpdate](),

	gateway.EventVoiceStateUpdate:  of[model.VoiceState](),
	gateway.EventVoiceServerUpdate: of[model.VoiceServerUpdate](),
}
