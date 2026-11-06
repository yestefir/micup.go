package gateway

const (
	EventReady       = "READY"
	EventResumed     = "RESUMED"
	EventRateLimited = "RATE_LIMITED"
)

const (
	EventGuildCreate = "GUILD_CREATE"
	EventGuildSync   = "GUILD_SYNC"
	EventGuildUpdate = "GUILD_UPDATE"
	EventGuildDelete = "GUILD_DELETE"
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
	EventInteractionCreate = "INTERACTION_CREATE"
)

const (
	EventVoiceStateUpdate  = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate = "VOICE_SERVER_UPDATE"
)

var knownEvents = []string{
	EventReady, EventResumed, EventRateLimited,
	EventGuildCreate, EventGuildSync, EventGuildUpdate, EventGuildDelete,
	EventGuildRoleCreate, EventGuildRoleUpdate, EventGuildRoleUpdateBulk, EventGuildRoleDelete,
	EventGuildMemberAdd, EventGuildMemberUpdate, EventGuildMemberRemove, EventGuildMembersChunk,
	EventGuildBanAdd, EventGuildBanRemove,
	EventChannelCreate, EventChannelUpdate, EventChannelUpdateBulk, EventChannelDelete, EventChannelPinsUpdate,
	EventMessageCreate, EventMessageUpdate, EventMessageDelete, EventMessageDeleteBulk,
	EventReactionAdd, EventReactionRemove, EventReactionRemoveAll, EventReactionRemoveEmoji,
	EventUserUpdate, EventPresenceUpdate, EventTypingStart,
	EventInviteCreate, EventInviteDelete, EventWebhooksUpdate,
	EventInteractionCreate,
	EventVoiceStateUpdate, EventVoiceServerUpdate,
}

func KnownEvents() []string { return append([]string(nil), knownEvents...) }
