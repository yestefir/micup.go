# Gateway events and replies

micup.go offers two ways to read Gateway events:

- `Client.NewSession(options).Run(ctx)` keeps a shard connected. It reconnects with bounded backoff, resumes the session when Micup allows it, keeps a guild/channel/voice cache, and joins voice channels. Most bots should use this.
- `Client.Gateway(ctx, handler)` holds one connection and returns when it ends. Use it when the application wants to own reconnect policy itself.

## Session

```go
session := client.NewSession(micup.SessionOptions{
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
		_, err := client.SendMessage(ctx, message.ChannelID, "pong", message.ID)
		return err
	},
	OnVoiceLost: func(guildID, connectionID string) {
		// Close the LiveKit room that belongs to this guild and connection.
	},
})
if err := session.Run(ctx); err != nil {
	return err
}
```

`Run` returns `nil` when the context ends. It returns a `*micup.CloseError` only when Micup closes the connection with a code that retrying cannot fix, such as an invalid token (`CloseError.Fatal()` reports this). Every other disconnect, reconnect request, heartbeat timeout, or handler error closes the current connection and starts a new one after `MinBackoff` (default 1 s), doubling up to `MaxBackoff` (default 15 s). The delay resets after a successful `READY` or `RESUMED`.

After a disconnect the session sends RESUME with the last session ID and sequence. An invalid-session response, or close code 4007 or 4009, clears that state so the next connection identifies again.

`Session.Ready()` is true between `READY`/`RESUMED` and the next disconnect. `Session.Send(opcode, data)` sends any native opcode on the live connection and returns `micup.ErrGatewayClosed` while disconnected.

The handler runs after the session updates its own state, so a handler that reads `Session.State()` sees the current event applied.

### State cache

`Session.State()` is updated from `READY`, `GUILD_CREATE`, `GUILD_DELETE`, `CHANNEL_CREATE`, `CHANNEL_UPDATE`, `CHANNEL_DELETE` and `VOICE_STATE_UPDATE`:

| Method | Returns |
|---|---|
| `GuildForChannel(channelID)` | The guild that owns a cached channel |
| `VoiceChannel(guildID, userID)` | The voice channel a user is in, or `""` |
| `BotVoiceChannel(guildID)` | The bot's own voice channel |
| `Guilds()`, `Channels()` | Cache sizes for diagnostics |

Native Micup messages can omit `guild_id`. `Session.ResolveGuild(ctx, channelID)` reads the cache and falls back to `GET /channels/{id}` once, then caches the answer. `Client.Channel(ctx, channelID)` performs that REST request directly.

### Voice channels

`Session.JoinVoice(ctx, guildID, channelID, micup.VoiceJoin{SelfDeaf: true})` sends opcode 4, waits for the matching `VOICE_SERVER_UPDATE`, and returns it as a `VoiceServerUpdate` with the LiveKit endpoint, short-lived token and connection ID. Bound the wait with the context. It returns:

- `ErrNotReady` before the session is ready,
- `ErrVoiceJoinPending` while another join for the same guild is waiting,
- `ErrVoiceGrant` when the update has no endpoint or token,
- `ErrVoiceMoved` when Micup grants a different channel.

`Session.Voice(guildID)` returns the current grant. `Session.LeaveVoice(guildID, connectionID)` sends the leave opcode only when the connection ID still matches the current grant, so a late leave cannot disconnect a newer connection.

`OnVoiceLost(guildID, connectionID)` runs when the bot is moved or disconnected from its voice channel, or when Micup sends an unrequested `VOICE_SERVER_UPDATE` that replaces a grant. The connection ID can be empty when no grant was stored. Close the matching LiveKit room. See [the voice guide](voice.md).

### Event names and payloads

Event names are exported as constants such as `micup.EventReady`, `micup.EventMessageCreate`, `micup.EventInteractionCreate` and `micup.EventVoiceServerUpdate`. `event.Decode(&value)` unmarshals the raw payload. `event.ShardID` is the shard that received the event.

## Shard manager

`micup.ShardForGuild(guildID, shardCount)` returns the shard that owns a guild, using the snowflake mapping `(guildID >> 22) % shardCount`.

`Client.NewShardManager` runs several shards in one process. Each shard is an ordinary `Session` with its own connection, resume state and cache; they share the handler and `OnVoiceLost` callback.

```go
manager, err := client.NewShardManager(micup.ShardManagerOptions{
	ShardCount: 4,
	Session: micup.SessionOptions{Handler: handle, OnVoiceLost: closeRoom},
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
| `ShardCount` | `Config.ShardCount` | Total shards across every process |
| `ShardIDs` | all of `0..ShardCount-1` | The shards this process runs |
| `IdentifyInterval` | `DefaultIdentifyInterval` (5 s) | Minimum gap between IDENTIFY requests from this manager |
| `Session` | | Options applied to every shard |

The manager waits `IdentifyInterval` between new identifies, including re-identifies after an invalid session. RESUME is not delayed. Split shards across processes with `ShardIDs`, for example `{0, 1}` in one process and `{2, 3}` in another with `ShardCount: 4`; each process then paces only its own identifies.

`Run` returns `nil` when the context ends. If any shard stops on a fatal close code, the manager stops every shard and returns that error wrapped with the shard ID; `errors.As(err, &closeErr)` still works. `Run` returns `ErrShardManagerRunning` if it is already running.

Routing helpers pick the owning shard from the guild ID:

| Method | Behavior |
|---|---|
| `ForGuild(guildID)` | The owning `*Session`, or `ErrShardNotOwned` when another process runs that shard |
| `JoinVoice`, `LeaveVoice`, `Voice` | The `Session` methods, sent on the owning shard |
| `ResolveGuild(ctx, channelID)` | Checks every local shard cache, then falls back to REST once |
| `Shard(id)`, `Shards()`, `ShardCount()`, `Ready()` | Inspection; `Ready` is true when every local shard is ready |

## Single connection

`Client.Gateway(ctx, handler)` sends IDENTIFY with the configured shard pair, answers heartbeat requests, and forwards dispatch packets as `Event` values. The handler receives the event name, sequence number, and raw JSON payload. `Event.Send` sends a native Gateway opcode while handling that event:

```go
payload := map[string]any{
	"guild_id": guildID,
	"channel_id": channelID,
	"self_mute": false,
	"self_deaf": true,
}
if err := event.Send(ctx, micup.OpVoiceStateUpdate, payload); err != nil {
	return err
}
```

The runner returns when the socket disconnects (`*CloseError` when a close code is known), when Micup asks for a reconnect (`ErrReconnectRequested`) or invalidates the session (`ErrSessionInvalidated`), or when the handler returns an error. It does not resume or keep a cache.

## Messages and interactions

Decode message payloads using `MessageCreate`. `Type` is `MessageTypeDefault` for a normal message and `MessageTypeReply` for a reply. Ignore messages authored by bots and use `Client.SendMessage` for plain replies or `Client.SendMessageWithOptions` for embeds. Automatic mentions are disabled for both methods.

Ephemeral messages are interaction responses, not regular channel messages. Decode an `INTERACTION_CREATE` event as `Interaction`, then answer it over REST with `RespondToInteraction`. Pass `true` for the `ephemeral` argument to make the response visible only to the user who invoked the interaction. Use `DeferInteraction` when work needs more time; the ephemeral choice is made when deferring and cannot be changed by later responses. A follow-up can be sent with `SendInteractionFollowup` while the interaction token remains valid.

```go
if event.Name == micup.EventInteractionCreate {
	var interaction micup.Interaction
	if err := event.Decode(&interaction); err != nil {
		return err
	}
	return client.RespondToInteraction(ctx, interaction, micup.MessageOptions{
		Embeds: []micup.Embed{{
			Title:       "Queue",
			Description: "Your track was added.",
			Color:       0x5865F2,
		}},
	}, true)
}
```

Do not run two sessions, managers or runners with the same shard ID unless the application has an explicit connection ownership policy.
