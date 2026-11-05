package gateway

import (
	"context"
	"encoding/json"
	"errors"
)

const (
	EventReady               = "READY"
	EventResumed             = "RESUMED"
	EventGuildCreate         = "GUILD_CREATE"
	EventGuildSync           = "GUILD_SYNC"
	EventGuildUpdate         = "GUILD_UPDATE"
	EventGuildDelete         = "GUILD_DELETE"
	EventGuildRoleCreate     = "GUILD_ROLE_CREATE"
	EventGuildRoleUpdate     = "GUILD_ROLE_UPDATE"
	EventGuildRoleUpdateBulk = "GUILD_ROLE_UPDATE_BULK"
	EventGuildRoleDelete     = "GUILD_ROLE_DELETE"
	EventGuildMemberAdd      = "GUILD_MEMBER_ADD"
	EventGuildMemberUpdate   = "GUILD_MEMBER_UPDATE"
	EventGuildMemberRemove   = "GUILD_MEMBER_REMOVE"
	EventGuildMembersChunk   = "GUILD_MEMBERS_CHUNK"
	EventGuildBanAdd         = "GUILD_BAN_ADD"
	EventGuildBanRemove      = "GUILD_BAN_REMOVE"
	EventChannelUpdateBulk   = "CHANNEL_UPDATE_BULK"
	EventChannelPinsUpdate   = "CHANNEL_PINS_UPDATE"
	EventMessageUpdate       = "MESSAGE_UPDATE"
	EventMessageDelete       = "MESSAGE_DELETE"
	EventMessageDeleteBulk   = "MESSAGE_DELETE_BULK"
	EventReactionAdd         = "MESSAGE_REACTION_ADD"
	EventReactionRemove      = "MESSAGE_REACTION_REMOVE"
	EventReactionRemoveAll   = "MESSAGE_REACTION_REMOVE_ALL"
	EventReactionRemoveEmoji = "MESSAGE_REACTION_REMOVE_EMOJI"
	EventTypingStart         = "TYPING_START"
	EventPresenceUpdate      = "PRESENCE_UPDATE"
	EventUserUpdate          = "USER_UPDATE"
	EventWebhooksUpdate      = "WEBHOOKS_UPDATE"
	EventInviteCreate        = "INVITE_CREATE"
	EventInviteDelete        = "INVITE_DELETE"
	EventRateLimited         = "RATE_LIMITED"
	EventChannelCreate       = "CHANNEL_CREATE"
	EventChannelUpdate       = "CHANNEL_UPDATE"
	EventChannelDelete       = "CHANNEL_DELETE"
	EventMessageCreate       = "MESSAGE_CREATE"
	EventInteractionCreate   = "INTERACTION_CREATE"
	EventVoiceStateUpdate    = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate   = "VOICE_SERVER_UPDATE"
)

type Event struct {
	Name     string
	Sequence *int64
	ShardID  int
	Data     json.RawMessage
	conn     *Conn
}

type Handler func(context.Context, Event) error

func (e Event) Decode(dest any) error { return json.Unmarshal(e.Data, dest) }

func (e Event) Send(ctx context.Context, opcode int, data any) error {
	if e.conn == nil {
		return errors.New("event is not attached to a Gateway connection")
	}
	return e.conn.Send(ctx, opcode, data)
}
