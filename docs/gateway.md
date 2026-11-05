# Gateway events, sessions and shards

micup.go offers three ways to read Gateway events, from most to least managed:

- `b.NewShardManager(options)` runs several shards in one process. See [Shard manager](#shard-manager).
- `b.NewSession(options)` keeps one shard connected. It reconnects with bounded backoff, resumes the session when Micup allows it, keeps a guild/channel/voice cache, and joins voice channels. Most single-shard bots use this.
- `b.Gateway(ctx, handler)` (or `micup.RunGateway`) holds one connection and returns when it ends. Use it when the application owns reconnect policy itself.

## Events

The `gateway` package defines `Event`, `Handler` and the event name constants: `EventReady`, `EventResumed`, `EventGuildCreate`, `EventGuildDelete`, `EventChannelCreate`, `EventChannelUpdate`, `EventChannelDelete`, `EventMessageCreate`, `EventInteractionCreate`, `EventVoiceStateUpdate`, `EventVoiceServerUpdate`. Any other name Micup sends is delivered as is.

| Field or method | Meaning |
|---|---|
| `Name` | Dispatch event name |
| `Sequence` | Gateway sequence number |
| `ShardID` | The shard that received the event |
| `Data` | Raw JSON payload |
| `Decode(&v)` | Unmarshals `Data` into a `model` type or your own struct |
| `Send(ctx, opcode, data)` | Sends a native opcode on the connection that delivered the event |

Opcodes are `micup.OpDispatch`, `OpHeartbeat`, `OpIdentify`, `OpVoiceStateUpdate`, `OpResume`, `OpReconnect`, `OpInvalidSession`, `OpHello` and `OpHeartbeatAck`.

## Session

```go
s, err := b.NewSession(micup.SessionOptions{
	Handler: func(ctx context.Context, event micup.Event) error {
		if event.Name != micup.EventMessageCreate {
			return nil
		}
		var message micup.MessageCreate
		if err := event.Decode(&message); err != nil {
			return err
		}
		if message.Author.Bot || message.Content != "!ping" {
			return nil
		}
		_, err := b.REST.SendMessage(ctx, message.ChannelID, "pong", message.ID)
		return err
	},
	OnVoiceLost: func(guildID, connectionID string) {
		// Close the LiveKit room that belongs to this guild and connection.
	},
})
if err != nil {
	return err
}
if err := s.Run(ctx); err != nil {
	return err
}
```

`Run` returns `nil` when the context ends.

A handler error is a request to see the event again: the session does not acknowledge an event until its handler returns `nil`, so after the reconnect Micup replays it with RESUME. Return an error only for failures worth retrying, and log and return `nil` for events you want to skip. It returns a `*micup.CloseError` only when Micup closes the connection with a code that retrying cannot fix, such as an invalid token (`CloseError.Fatal()` reports this). Every other disconnect, reconnect request, heartbeat timeout, or handler error closes the current connection and starts a new one after `MinBackoff` (default `micup.DefaultMinBackoff`, 1 s), doubling up to `MaxBackoff` (default 15 s). The delay resets after a successful `READY` or `RESUMED`.

After a disconnect the session sends RESUME with the last session ID and the last sequence it finished processing. Heartbeats carry the same processed sequence, because Micup drops replayable events up to it. An invalid-session response (opcode 9) clears the resume state and the session identifies again on the same socket, paced by the shard manager's identify limiter. Close codes 4007 and 4009 (`CloseError.ResetsSession()`) clear it before the next connection.

`Session.Ready()` is true between `READY`/`RESUMED` and the next disconnect. `Session.Send(opcode, data)` sends any native opcode on the live connection and returns `micup.ErrGatewayClosed` while disconnected. `Session.ShardID()` returns the shard it runs.

The handler runs after the session updates its own state, so a handler that reads `Session.State()` sees the current event applied.

Without `bot`, build a session with `micup.NewSession(micup.SessionConfig{Endpoint, Token, BotID, ShardID, ShardCount, Channels}, options)`. `Channels` is any `micup.ChannelLookup`; `*micup.RESTClient` satisfies it.

### State cache

`Session.State()` returns a `*micup.State` updated from `READY`, `GUILD_CREATE`, `GUILD_DELETE`, `CHANNEL_CREATE`, `CHANNEL_UPDATE`, `CHANNEL_DELETE` and `VOICE_STATE_UPDATE`:

| Method | Returns |
|---|---|
| `GuildForChannel(channelID)` | The guild that owns a cached channel |
| `VoiceChannel(guildID, userID)` | The voice channel a user is in, or `""` |
| `BotVoiceChannel(guildID)` | The bot's own voice channel |
| `Guilds()`, `Channels()` | Cache sizes for diagnostics |
| `Apply(name, data)` | Applies an event; returns a `*micup.VoiceMove` when the bot's own voice channel changed |
| `Remember(channelID, guildID)` | Adds a channel learned elsewhere |

Native Micup messages can omit `guild_id`. `Session.ResolveGuild(ctx, channelID)` reads the cache and falls back to `GET /channels/{id}` once, then caches the answer. `b.REST.Channel(ctx, channelID)` performs that request directly.

### Voice channels

`Session.JoinVoice(ctx, guildID, channelID, micup.VoiceJoin{SelfDeaf: true})` sends opcode 4, waits for the matching `VOICE_SERVER_UPDATE`, and returns it as a `micup.VoiceServerUpdate` with the LiveKit endpoint, short-lived token and connection ID. Bound the wait with the context. It returns:

- `micup.ErrNotReady` before the session is ready,
- `micup.ErrVoiceJoinPending` while another join for the same guild is waiting,
- `micup.ErrVoiceGrant` when the update has no endpoint or token,
- `micup.ErrVoiceMoved` when Micup grants a different channel,
- a `*micup.VoiceRefusedError` (matching `errors.Is(err, micup.ErrVoiceRefused)`) when Micup refuses the join with opcode 12. `Code` is `micup.VoiceChannelFull`, `micup.VoiceP2PConsentRequired` or `micup.VoiceP2PUnavailable`.

Micup reports only those three refusals. Other refusals, such as a missing CONNECT permission, send nothing, so always bound `JoinVoice` with a context deadline. Opcode 12 carries no guild ID; when several joins are pending, the session assigns a refusal to the oldest one.

`Session.Voice(guildID)` returns the current grant. `Session.LeaveVoice(guildID, connectionID)` sends the leave opcode only when the connection ID still matches the current grant, so a late leave cannot disconnect a newer connection.

`OnVoiceLost(guildID, connectionID)` runs when the bot is disconnected from its voice channel, moved to a channel it holds no grant for, or loses its session (invalid session or a new READY). The connection ID can be empty when no grant was stored. Close the matching LiveKit room. See [the voice guide](voice.md).

Micup sends a new grant without a request when the connection moves to another region or channel. With `OnVoiceGrant(grant)` set, the session stores that grant and passes it to the callback: reconnect the LiveKit room to `grant.Endpoint` with `grant.Token` and keep playing. Without `OnVoiceGrant`, an unrequested grant that replaces a stored one is reported through `OnVoiceLost`, as in earlier releases.

## Shard manager

`micup.ShardForGuild(guildID, shardCount)` returns the shard that owns a guild, using the snowflake mapping `(guildID >> 22) % shardCount`.

`b.NewShardManager` (or `micup.NewShardManager(b.SessionConfig(), options)`) runs several shards in one process. Each shard is an ordinary `*micup.Session` with its own connection, resume state and cache; they share the handler and `OnVoiceLost` callback.

```go
manager, err := b.NewShardManager(micup.ShardOptions{
	ShardCount: 4,
	Session:    micup.SessionOptions{Handler: handle, OnVoiceLost: closeRoom},
})
if err != nil {
	return err
}
if err := manager.Run(ctx); err != nil {
	return err
}
```

| Option | Default | Meaning |
|---|---|---|
| `ShardCount` | `micup.Config.ShardCount` | Total shards across every process |
| `ShardIDs` | all of `0..ShardCount-1` | The shards this process runs |
| `IdentifyInterval` | `micup.DefaultIdentifyInterval` (5 s) | Minimum gap between IDENTIFY requests from this manager |
| `Session` | | Options applied to every shard |

The manager waits `IdentifyInterval` between new identifies, including re-identifies after an invalid session. RESUME is not delayed. Split shards across processes with `ShardIDs`, for example `{0, 1}` in one process and `{2, 3}` in another with `ShardCount: 4`; each process then paces only its own identifies.

`Run` returns `nil` when the context ends.

A handler error is a request to see the event again: the session does not acknowledge an event until its handler returns `nil`, so after the reconnect Micup replays it with RESUME. Return an error only for failures worth retrying, and log and return `nil` for events you want to skip. If any shard stops on a fatal close code, the manager stops every shard and returns that error wrapped with the shard ID; `errors.As(err, &closeErr)` still works. `Run` returns `micup.ErrShardManagerRunning` if it is already running.

| Method | Behavior |
|---|---|
| `ForGuild(guildID)` | The owning `*micup.Session`, or `micup.ErrShardNotOwned` when another process runs that shard |
| `JoinVoice`, `LeaveVoice`, `Voice` | The `Session` methods, sent on the owning shard |
| `ResolveGuild(ctx, channelID)` | Checks every local shard cache, then falls back to REST once |
| `Shard(id)`, `Shards()`, `ShardCount()`, `Ready()` | Inspection; `Ready` is true when every local shard is ready |

## Single connection

`b.Gateway(ctx, handler)` runs `micup.RunGateway` once with the configured shard pair. It sends IDENTIFY, answers heartbeat requests, and forwards dispatch packets as `Event` values:

```go
payload := map[string]any{
	"guild_id":   guildID,
	"channel_id": channelID,
	"self_mute":  false,
	"self_deaf":  true,
}
if err := event.Send(ctx, micup.OpVoiceStateUpdate, payload); err != nil {
	return err
}
```

It returns when the socket disconnects (`*micup.CloseError` when a close code is known), when Micup asks for a reconnect (`micup.ErrReconnectRequested`) or invalidates the session (`micup.ErrSessionInvalidated`), or when the handler returns an error. It does not resume or keep a cache unless you pass your own `micup.GatewayOptions.Resume` between calls.

## Messages and interactions

Decode message payloads into `micup.MessageCreate`. `Type` is `micup.MessageTypeDefault` for a normal message and `micup.MessageTypeReply` for a reply. Ignore messages authored by bots and use `micup.RESTClient.SendMessage` for plain replies or `SendMessageWithOptions` for embeds. Automatic mentions are disabled for both methods.

Ephemeral messages are interaction responses, not regular channel messages. Decode an `INTERACTION_CREATE` event as `micup.Interaction`, then answer it over REST with `RespondToInteraction`. Pass `true` for the `ephemeral` argument to make the response visible only to the user who invoked the interaction. Use `DeferInteraction` when work needs more time; the ephemeral choice is made when deferring and cannot be changed by later responses. A follow-up can be sent with `SendInteractionFollowup` while the interaction token remains valid.

```go
if event.Name == micup.EventInteractionCreate {
	var interaction micup.Interaction
	if err := event.Decode(&interaction); err != nil {
		return err
	}
	return b.REST.RespondToInteraction(ctx, interaction, micup.MessageOptions{
		Embeds: []micup.Embed{{
			Title:       "Queue",
			Description: "Your track was added.",
			Color:       0x5865F2,
		}},
	}, true)
}
```

Do not run two sessions, managers or runners with the same shard ID unless the application has an explicit connection ownership policy.
