package gateway

const (
	EventReady                = "READY"
	EventResumed              = "RESUMED"
	EventRateLimited          = "RATE_LIMITED"
	EventSessionsReplace      = "SESSIONS_REPLACE"
	EventGatewayRolloutUpdate = "GATEWAY_ROLLOUT_UPDATE"
)

const (
	EventGuildCreate              = "GUILD_CREATE"
	EventGuildSync                = "GUILD_SYNC"
	EventGuildUpdate              = "GUILD_UPDATE"
	EventGuildDelete              = "GUILD_DELETE"
	EventGuildAuditLogEntryCreate = "GUILD_AUDIT_LOG_ENTRY_CREATE"
)

const (
	EventGuildRoleCreate     = "GUILD_ROLE_CREATE"
	EventGuildRoleUpdate     = "GUILD_ROLE_UPDATE"
	EventGuildRoleUpdateBulk = "GUILD_ROLE_UPDATE_BULK"
	EventGuildRoleDelete     = "GUILD_ROLE_DELETE"
)

const (
	EventGuildMemberAdd    = "GUILD_MEMBER_ADD"
	EventGuildMemberUpdate = "GUILD_MEMBER_UPDATE"
	EventGuildMemberRemove = "GUILD_MEMBER_REMOVE"
	EventGuildMembersChunk = "GUILD_MEMBERS_CHUNK"
	EventGuildBanAdd       = "GUILD_BAN_ADD"
	EventGuildBanRemove    = "GUILD_BAN_REMOVE"
)

const (
	EventChannelCreate     = "CHANNEL_CREATE"
	EventChannelUpdate     = "CHANNEL_UPDATE"
	EventChannelUpdateBulk = "CHANNEL_UPDATE_BULK"
	EventChannelDelete     = "CHANNEL_DELETE"
	EventChannelPinsUpdate = "CHANNEL_PINS_UPDATE"
)

const (
	EventMessageCreate     = "MESSAGE_CREATE"
	EventMessageUpdate     = "MESSAGE_UPDATE"
	EventMessageDelete     = "MESSAGE_DELETE"
	EventMessageDeleteBulk = "MESSAGE_DELETE_BULK"
)

const (
	EventReactionAdd         = "MESSAGE_REACTION_ADD"
	EventReactionRemove      = "MESSAGE_REACTION_REMOVE"
	EventReactionRemoveAll   = "MESSAGE_REACTION_REMOVE_ALL"
	EventReactionRemoveEmoji = "MESSAGE_REACTION_REMOVE_EMOJI"
)

const (
	EventUserUpdate     = "USER_UPDATE"
	EventPresenceUpdate = "PRESENCE_UPDATE"
	EventTypingStart    = "TYPING_START"
)

const (
	EventInviteCreate   = "INVITE_CREATE"
	EventInviteDelete   = "INVITE_DELETE"
	EventWebhooksUpdate = "WEBHOOKS_UPDATE"
)

const (
	EventInteractionCreate         = "INTERACTION_CREATE"
	EventApplicationCommandsUpdate = "APPLICATION_COMMANDS_UPDATE"
)

const (
	EventVoiceStateUpdate  = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate = "VOICE_SERVER_UPDATE"
)

const (
	EventGuildEmojisUpdate         = "GUILD_EMOJIS_UPDATE"
	EventGuildStickersUpdate       = "GUILD_STICKERS_UPDATE"
	EventGuildMemberListUpdate     = "GUILD_MEMBER_LIST_UPDATE"
	EventGuildCountsUpdate         = "GUILD_COUNTS_UPDATE"
	EventChannelRecipientAdd       = "CHANNEL_RECIPIENT_ADD"
	EventChannelRecipientRemove    = "CHANNEL_RECIPIENT_REMOVE"
	EventChannelMemberCountsUpdate = "CHANNEL_MEMBER_COUNTS_UPDATE"
	EventPresenceUpdateBulk        = "PRESENCE_UPDATE_BULK"
	EventPassiveUpdates            = "PASSIVE_UPDATES"
	EventReactionAddMany           = "MESSAGE_REACTION_ADD_MANY"
)

const (
	EventThreadCreate           = "THREAD_CREATE"
	EventThreadUpdate           = "THREAD_UPDATE"
	EventThreadDelete           = "THREAD_DELETE"
	EventThreadListSync         = "THREAD_LIST_SYNC"
	EventThreadMemberUpdate     = "THREAD_MEMBER_UPDATE"
	EventThreadMembersUpdate    = "THREAD_MEMBERS_UPDATE"
	EventThreadMemberListUpdate = "THREAD_MEMBER_LIST_UPDATE"
)

const (
	EventCallCreate         = "CALL_CREATE"
	EventCallUpdate         = "CALL_UPDATE"
	EventCallDelete         = "CALL_DELETE"
	EventEntranceSoundPlay  = "ENTRANCE_SOUND_PLAY"
	EventRelationshipAdd    = "RELATIONSHIP_ADD"
	EventRelationshipUpdate = "RELATIONSHIP_UPDATE"
	EventRelationshipRemove = "RELATIONSHIP_REMOVE"
)

var knownEvents = []string{
	EventReady, EventResumed, EventRateLimited, EventSessionsReplace, EventGatewayRolloutUpdate,
	EventGuildCreate, EventGuildSync, EventGuildUpdate, EventGuildDelete, EventGuildAuditLogEntryCreate,
	EventGuildRoleCreate, EventGuildRoleUpdate, EventGuildRoleUpdateBulk, EventGuildRoleDelete,
	EventGuildMemberAdd, EventGuildMemberUpdate, EventGuildMemberRemove, EventGuildMembersChunk,
	EventGuildBanAdd, EventGuildBanRemove,
	EventChannelCreate, EventChannelUpdate, EventChannelUpdateBulk, EventChannelDelete, EventChannelPinsUpdate,
	EventMessageCreate, EventMessageUpdate, EventMessageDelete, EventMessageDeleteBulk,
	EventReactionAdd, EventReactionRemove, EventReactionRemoveAll, EventReactionRemoveEmoji,
	EventUserUpdate, EventPresenceUpdate, EventTypingStart,
	EventInviteCreate, EventInviteDelete, EventWebhooksUpdate,
	EventInteractionCreate, EventApplicationCommandsUpdate,
	EventVoiceStateUpdate, EventVoiceServerUpdate,
	EventGuildEmojisUpdate, EventGuildStickersUpdate, EventGuildMemberListUpdate, EventGuildCountsUpdate,
	EventChannelRecipientAdd, EventChannelRecipientRemove, EventChannelMemberCountsUpdate,
	EventPresenceUpdateBulk, EventPassiveUpdates, EventReactionAddMany,
	EventThreadCreate, EventThreadUpdate, EventThreadDelete, EventThreadListSync,
	EventThreadMemberUpdate, EventThreadMembersUpdate, EventThreadMemberListUpdate,
	EventCallCreate, EventCallUpdate, EventCallDelete, EventEntranceSoundPlay,
	EventRelationshipAdd, EventRelationshipUpdate, EventRelationshipRemove,
}

func KnownEvents() []string { return append([]string(nil), knownEvents...) }
