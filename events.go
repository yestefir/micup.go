package micup

import (
	"context"
	"encoding/json"
	"errors"
)

const (
	EventReady             = "READY"
	EventResumed           = "RESUMED"
	EventGuildCreate       = "GUILD_CREATE"
	EventGuildDelete       = "GUILD_DELETE"
	EventChannelCreate     = "CHANNEL_CREATE"
	EventChannelUpdate     = "CHANNEL_UPDATE"
	EventChannelDelete     = "CHANNEL_DELETE"
	EventMessageCreate     = "MESSAGE_CREATE"
	EventInteractionCreate = "INTERACTION_CREATE"
	EventVoiceStateUpdate  = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate = "VOICE_SERVER_UPDATE"
)

type Event struct {
	Name     string
	Sequence *int64
	ShardID  int
	Data     json.RawMessage
	session  *gatewaySession
}

type EventHandler func(context.Context, Event) error

func (e Event) Send(ctx context.Context, opcode int, data any) error {
	if e.session == nil {
		return errors.New("event is not attached to a Gateway session")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return e.session.send(opcode, data)
}

func (e Event) Decode(dest any) error { return json.Unmarshal(e.Data, dest) }
