# Gateway events, sessions and shards

micup.go offers three ways to read Gateway events, from most to least managed:

- `b.NewShardManager(options)` runs several shards in one process. See [Shard manager](#shard-manager).
- `b.NewSession(options)` keeps one shard connected. It reconnects with bounded backoff, resumes the session when Micup allows it, keeps a guild/channel/voice cache, and joins voice channels. Most single-shard bots use this.
- `b.Gateway(ctx, handler)` (or `micup.RunGateway`) holds one connection and returns when it ends. Use it when the application owns reconnect policy itself.

## Events

Handlers receive a `micup.Event`. Every event name, its payload type, typed decoding and the event mux are described in [Events and typed payloads](events.md). Any name Micup sends is delivered, including names the module does not know yet.

| Field or method | Meaning |
|---|---|
| `Name` | Dispatch event name |
| `Sequence` | Gateway sequence number |
| `ShardID` | The shard that received the event |
| `Data` | Raw JSON payload |
| `Decode(&v)` | Unmarshals `Data` into a `model` type or your own struct |
| `Send(ctx, opcode, data)` | Sends a native opcode on the connection that delivered the event |

Use `micup.EventAs[T](event)` or a `micup.NewEventMux()` for typed payloads; see [Events and typed payloads](events.md).

Opcodes are `micup.OpDispatch`, `OpHeartbeat`, `OpIdentify`, `OpPresenceUpdate`, `OpVoiceStateUpdate`, `OpResume`, `OpReconnect`, `OpRequestGuildMembers`, `OpInvalidSession`, `OpHello`, `OpHeartbeatAck`, `OpGatewayError`, `OpLazyRequest`, `OpRequestGuildCounts` and `OpRequestChannelMemberCounts`. Bots need no Lazy Request; Micup never treats bot sessions as passive.

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

### Identify options and presence

| `SessionOptions` field | Effect |
|---|---|
| `IgnoredEvents` | Dispatch names Micup should not send this session (up to 256), for example `TYPING_START`, `MESSAGE_REACTION_ADD`, `MESSAGE_REACTION_REMOVE`, `PRESENCE_UPDATE`. Micup has no intents; this is the way to cut traffic. Ignored events use no sequence number. `MESSAGE_CREATE` still arrives for messages that mention the bot directly |
| `Presence` | The status published when the session starts: `micup.Presence{Status: micup.StatusOnline, CustomStatus: &micup.CustomStatus{Text: "!help"}}` |
| `Flags` | Session flags: `micup.FlagChannelThreads`, `micup.FlagDebounceMessageReactions`, `micup.FlagGuildVisits` |
| `Compress` | Connects with `compress=zstd-stream&stream=1`. Every server frame is then one part of a single zstd stream, and every client message goes through one compressor for the whole connection, as Micup requires. Cuts Gateway bandwidth substantially for bots in many guilds; off by default |

Session construction copies `IgnoredEvents`. Reusing or editing the caller's slice afterward does not change an existing session's event filtering or later IDENTIFY payloads.

#### Changing the bot's status

Each call sends opcode 3 and remembers the result, so the next IDENTIFY publishes it too. `ShardManager` has the same methods and applies them to every shard. A shard that is disconnected keeps the new presence and sends it when it identifies again.

| Method | Effect |
|---|---|
| `SetStatus(ctx, micup.StatusIdle)` | Changes the status and keeps the custom status |
| `SetCustomStatus(ctx, micup.CustomStatus{Text: "!help", EmojiName: "🎵"})` | Changes the custom status and keeps the status |
| `ClearCustomStatus(ctx)` | Removes the custom status |
| `UpdatePresence(ctx, presence)` | Replaces the whole presence. A nil `CustomStatus` clears the custom status |
| `Presence()` | The presence the session publishes; `online` with no custom status before any change |

The session copies custom status values supplied through initial options or `UpdatePresence`. `Presence()` returns an independent snapshot, including its custom status; editing it does not change the session. Use the update methods to publish a change.

```go
status := micup.CustomStatus{Text: "Bakımdayız", EmojiName: "🔧"}.Until(time.Now().Add(30 * time.Minute))
if err := shards.SetCustomStatus(ctx, status); err != nil {
	return err
}
```

Status is one of `online`, `idle`, `dnd`, `invisible` or `offline`; Micup publishes `offline` as `invisible`. A custom status needs text or an emoji. Text is at most `micup.MaxCustomStatusText` (128) and `EmojiName` at most `micup.MaxCustomStatusEmoji` (32) UTF-16 units, as Micup counts them. `EmojiName` must be a single Unicode emoji, and Micup ignores it when `EmojiID` is set. `CustomStatus.Until(t)` sets `ExpiresAt`, and Micup clears the status at that time; it must be in the future.

Micup drops an invalid presence update without telling the bot. The methods and `SessionOptions.Presence` therefore check these limits first, return an error, and send nothing. Micup's checks that need its data, such as whether a custom emoji exists or content moderation, cannot be done here.

Micup limits each connection to 600 commands per 60 seconds and 5 presence updates per 20 seconds, and closes or drops over the limit. The connection queues commands at 540 per 60 seconds (heartbeats, IDENTIFY and RESUME are not counted against that) and presence updates at 5 per 20 seconds, so `Send` and the presence methods wait instead of being dropped; the wait ends with the context. A command over Micup's 16,384-byte limit returns an error instead of being sent.

`Session.Latency()` is the round trip of the last acknowledged heartbeat on the live connection. It is 0 while disconnected and until the first acknowledgement, which comes one heartbeat interval (about 41 seconds on Micup) after connecting.

`Session.Ready()` is true between `READY`/`RESUMED` and the next disconnect. `Session.Send(ctx, opcode, data)` sends any native opcode on the live connection and returns `micup.ErrGatewayClosed` while disconnected. `Session.ShardID()` returns the shard it runs.

The handler runs after the session updates its own state, so a handler that reads `Session.State()` sees the current event applied.

Without `bot`, build a session with `micup.NewSession(micup.SessionConfig{Endpoint, Token, BotID, ShardID, ShardCount, Channels}, options)`. `Channels` is any `micup.ChannelLookup`; `*micup.RESTClient` satisfies it.

### State cache

`Session.State()` returns a `*micup.State` built from `READY`, `GUILD_CREATE`, `GUILD_SYNC`, `GUILD_UPDATE`, `GUILD_DELETE`, the role, member and channel events (including `CHANNEL_UPDATE_BULK` and `GUILD_ROLE_UPDATE_BULK`), `MESSAGE_CREATE` (the author's `member`), `GUILD_MEMBERS_CHUNK` and `VOICE_STATE_UPDATE`. `GUILD_CREATE` and `GUILD_SYNC` replace a guild's roles, channels and voice states and add their members, as Micup specifies; `GUILD_DELETE` with `unavailable: true` keeps the guild and marks it unavailable.

| Method | Returns |
|---|---|
| `GuildForChannel(channelID)`, `Channel(channelID)` | The owning guild, or the cached channel with its permission overwrites |
| `Role(guildID, roleID)`, `MemberRoles(guildID, userID)`, `OwnerID(guildID)` | Cached roles, a member's role IDs and the owner |
| `Permissions(guildID, channelID, userID)`, `BotPermissions(guildID, channelID)` | The computed permission mask and whether the cache knew enough to compute it |
| `VoiceChannel(guildID, userID)`, `VoiceMembers(guildID, channelID)`, `BotVoiceChannel(guildID)` | Voice state |
| `Unavailable(guildID)`, `Guilds()`, `Channels()` | Outage flag and cache sizes |
| `Apply(name, data)` | Applies an event; returns a `*micup.VoiceMove` when the bot's own voice channel changed |
| `Remember(channelID, guildID)` | Adds a channel learned elsewhere |

`Channel` and `Role` return independent snapshots, including nested pointers, slices and maps. Editing a returned snapshot does not change cached state or another caller's snapshot.

Micup sends only the bot's own member and the members in voice channels with `GUILD_CREATE`, so `Permissions` is known for the bot and voice users and grows as messages, member events and chunks arrive. It follows Micup's order: the owner has every permission, `ADMINISTRATOR` skips overwrites, then the `@everyone` overwrite, the member's role overwrites combined (an allow beats a deny), and the member's own overwrite. `micup.ComputePermissions` runs the same computation on your own data, and `Permissions.Has(micup.PermissionConnect | micup.PermissionSpeak)` checks bits.

### Member requests

`Session.RequestGuildMembers(ctx, micup.MembersRequest{GuildID, Query, Limit, UserIDs, Presences})` sends opcode 8 with a nonce, collects every `GUILD_MEMBERS_CHUNK` page and returns the members, which are also added to the cache. An empty `Query` with `Limit` 0 asks for the full list, which needs `MANAGE_ROLES`, `KICK_MEMBERS` and `BAN_MEMBERS` and is allowed once every 30 seconds per guild; a refused request returns `*micup.MembersRateLimitedError` (`errors.Is(err, micup.ErrMembersRateLimited)`) with `RetryAfter`. Micup drops a request without permission silently, so bound the call with a context. Micup also silently drops an opcode 8 request when two are already being answered on the connection, so a session keeps at most two requests in flight and queues the rest; a queued call still ends with its context. `ShardManager.RequestGuildMembers` routes by guild.

If member responses arrive faster than the request can consume its bounded buffer, the call returns an error with any members already collected. Treat those results as incomplete. Other requests and Gateway event processing continue.

When `Session.Run` stops, pending member requests are notified with `micup.ErrGatewayClosed`, even if their own contexts remain active. Already collected members may accompany that error. Transient reconnects keep `Run` active and do not trigger this shutdown cleanup.

Native Micup messages can omit `guild_id`. `Session.ResolveGuild(ctx, channelID)` reads the cache and falls back to `GET /channels/{id}` once, then caches the answer. `b.REST.Channel(ctx, channelID)` performs that request directly.

### Voice channels

`Session.JoinVoice(ctx, guildID, channelID, micup.VoiceJoin{SelfDeaf: true})` sends opcode 4, waits for the matching `VOICE_SERVER_UPDATE`, and returns it as a `micup.VoiceServerUpdate` with the LiveKit endpoint, short-lived token and connection ID. Bound the wait with the context. It returns:

- `micup.ErrNotReady` before the session is ready,
- `micup.ErrGatewayClosed` when the session's `Run` stops while the join is pending,
- `micup.ErrVoicePermission` when the cache shows the bot lacks `VIEW_CHANNEL` or `CONNECT` in the channel (Micup would ignore the join silently),
- `micup.ErrVoiceJoinPending` while another join for the same guild is waiting,
- `micup.ErrVoiceGrant` when the update has no endpoint or token,
- `micup.ErrVoiceMoved` when Micup grants a different channel,
- a `*micup.VoiceRefusedError` (matching `errors.Is(err, micup.ErrVoiceRefused)`) when Micup refuses the join with opcode 12. `Code` is `micup.VoiceChannelFull`, `micup.VoiceP2PConsentRequired` or `micup.VoiceP2PUnavailable`.

Micup reports only those three refusals. Other refusals, such as a missing CONNECT permission, send nothing, so always bound `JoinVoice` with a context deadline. Opcode 12 carries no guild ID; when several joins are pending, the session assigns a refusal to the oldest one.

Pending joins are notified when `Run` ends even if their own caller contexts remain active. Transient reconnects keep `Run` active and retain the existing join behavior.

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

Micup allows at most 2,500 guilds per bot shard and closes a shard with more with 4011 (`Sharding required`, fatal). Raise `ShardCount` until every shard is below that; Micup accepts 1 to 16,384 shards. `GET /v1/gateway/bot` always recommends 1 shard on Micup, so the module does not ask it for a count. Events without a guild (DMs, account events) arrive only on shard 0.

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

It returns when the socket disconnects (`*micup.CloseError` when a close code is known), when Micup asks for a reconnect (`micup.ErrReconnectRequested`), or when the handler returns an error. An invalid session is handled on the same socket by identifying again. It does not resume or keep a cache unless you pass your own `micup.GatewayOptions.Resume` between calls.

## Messages and interactions

Decode message payloads into `micup.MessageCreate`. `Type` is `micup.MessageTypeDefault` for a normal message and `micup.MessageTypeReply` for a reply. Ignore messages authored by bots and use `micup.RESTClient.SendMessage` for plain replies or `SendMessageWithOptions` for embeds. Automatic mentions are disabled for both methods.

Bots receive an `INTERACTION_CREATE` event when someone presses one of their buttons. [Buttons](components.md) covers sending buttons, reading clicks and every way to answer them.

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
