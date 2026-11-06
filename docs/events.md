# Events and typed payloads

Every Gateway dispatch arrives as a `micup.Event` with a `Name` and raw JSON `Data`. micup.go knows a payload type for each event in the table below. Names it does not know are still delivered unchanged, so new Micup events reach the handler before the module learns them.

## Reading a payload

`micup.EventAs[T]` decodes an event into its payload type. It returns an error when `T` is not the type registered for that event name, so a handler cannot silently read a `MESSAGE_CREATE` as a `MessageDelete`.

```go
switch event.Name {
case micup.EventMessageDelete:
	deleted, err := micup.EventAs[micup.MessageDelete](event)
	if err != nil {
		return err
	}
	log.Printf("message %s deleted in %s", deleted.ID, deleted.ChannelID)
}
```

`micup.DecodeEvent(event)` returns the payload as `any` for a type switch. For an unknown name it returns the raw `json.RawMessage` and no error.

`micup.KnownEvents()` lists every event name the module defines, grouped by domain.

## Event mux

`micup.NewEventMux()` routes events to typed listeners. `mux.Handle` is a `micup.EventHandler`, so it plugs into `SessionOptions.Handler`, `ShardOptions.Handler` or `b.Gateway`.

```go
mux := micup.NewEventMux()
mux.On(micup.EventMessageCreate, micup.Listen(func(ctx context.Context, m micup.MessageCreate) error {
	if m.Author.Bot || m.Content != "!ping" {
		return nil
	}
	_, err := b.REST.SendMessage(ctx, m.ChannelID, "pong", m.ID)
	return err
}))
mux.On(micup.EventGuildRoleCreate, micup.Listen(func(ctx context.Context, e micup.RoleEvent) error {
	log.Printf("role %s created in %s", e.Role.Name, e.GuildID)
	return nil
}))
mux.HandleRaw("SOMETHING_NEW", func(ctx context.Context, event micup.Event) error { return nil })
mux.HandleAll(func(ctx context.Context, event micup.Event) error { return nil })

s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
```

| Method | Behavior |
|---|---|
| `On(name, micup.Listen(fn))` | Calls `fn` with the decoded payload. Panics at registration when `fn`'s parameter type is not the payload type of `name` |
| `HandleRaw(name, handler)` | Calls a plain `EventHandler` for one name, including names without a payload type |
| `HandleAll(handler)` | Calls `handler` for every event, after the named handlers |
| `Handle(ctx, event)` | Runs every matching handler in registration order and returns their errors joined |

Handlers for one event run in order on the session's goroutine; a slow handler delays the next event.

## Payload types

The table is generated from the registry. Captured means the payload type is tested against an anonymized payload recorded from Micup.

<!-- events:start -->
| Event | Constant | Payload | Captured |
|---|---|---|---|
| `READY` | `micup.EventReady` | `micup.Ready` | yes |
| `RESUMED` | `micup.EventResumed` | `micup.Resumed` | no |
| `RATE_LIMITED` | `micup.EventRateLimited` | `micup.RateLimited` | no |
| `SESSIONS_REPLACE` | `micup.EventSessionsReplace` | `micup.SessionsReplace` | yes |
| `GUILD_CREATE` | `micup.EventGuildCreate` | `micup.GuildCreate` | yes |
| `GUILD_SYNC` | `micup.EventGuildSync` | `micup.GuildCreate` | no |
| `GUILD_UPDATE` | `micup.EventGuildUpdate` | `micup.GuildUpdate` | yes |
| `GUILD_DELETE` | `micup.EventGuildDelete` | `micup.GuildDelete` | no |
| `GUILD_AUDIT_LOG_ENTRY_CREATE` | `micup.EventGuildAuditLogEntryCreate` | `micup.AuditLogEntryCreate` | yes |
| `GUILD_ROLE_CREATE` | `micup.EventGuildRoleCreate` | `micup.RoleEvent` | yes |
| `GUILD_ROLE_UPDATE` | `micup.EventGuildRoleUpdate` | `micup.RoleEvent` | yes |
| `GUILD_ROLE_UPDATE_BULK` | `micup.EventGuildRoleUpdateBulk` | `micup.RoleUpdateBulk` | yes |
| `GUILD_ROLE_DELETE` | `micup.EventGuildRoleDelete` | `micup.RoleDelete` | yes |
| `GUILD_MEMBER_ADD` | `micup.EventGuildMemberAdd` | `micup.GuildMember` | no |
| `GUILD_MEMBER_UPDATE` | `micup.EventGuildMemberUpdate` | `micup.GuildMember` | yes |
| `GUILD_MEMBER_REMOVE` | `micup.EventGuildMemberRemove` | `micup.GuildMemberRemove` | no |
| `GUILD_MEMBERS_CHUNK` | `micup.EventGuildMembersChunk` | `micup.MembersChunk` | yes |
| `GUILD_BAN_ADD` | `micup.EventGuildBanAdd` | `micup.GuildBan` | no |
| `GUILD_BAN_REMOVE` | `micup.EventGuildBanRemove` | `micup.GuildBan` | no |
| `CHANNEL_CREATE` | `micup.EventChannelCreate` | `micup.Channel` | yes |
| `CHANNEL_UPDATE` | `micup.EventChannelUpdate` | `micup.Channel` | yes |
| `CHANNEL_UPDATE_BULK` | `micup.EventChannelUpdateBulk` | `micup.ChannelUpdateBulk` | yes |
| `CHANNEL_DELETE` | `micup.EventChannelDelete` | `micup.Channel` | yes |
| `CHANNEL_PINS_UPDATE` | `micup.EventChannelPinsUpdate` | `micup.ChannelPinsUpdate` | yes |
| `MESSAGE_CREATE` | `micup.EventMessageCreate` | `micup.MessageCreate` | yes |
| `MESSAGE_UPDATE` | `micup.EventMessageUpdate` | `micup.MessageCreate` | yes |
| `MESSAGE_DELETE` | `micup.EventMessageDelete` | `micup.MessageDelete` | yes |
| `MESSAGE_DELETE_BULK` | `micup.EventMessageDeleteBulk` | `micup.MessageDeleteBulk` | yes |
| `MESSAGE_REACTION_ADD` | `micup.EventReactionAdd` | `micup.ReactionEvent` | yes |
| `MESSAGE_REACTION_REMOVE` | `micup.EventReactionRemove` | `micup.ReactionEvent` | yes |
| `MESSAGE_REACTION_REMOVE_ALL` | `micup.EventReactionRemoveAll` | `micup.ReactionRemoveAll` | yes |
| `MESSAGE_REACTION_REMOVE_EMOJI` | `micup.EventReactionRemoveEmoji` | `micup.ReactionRemoveEmoji` | yes |
| `USER_UPDATE` | `micup.EventUserUpdate` | `micup.User` | yes |
| `PRESENCE_UPDATE` | `micup.EventPresenceUpdate` | `micup.PresenceUpdate` | no |
| `TYPING_START` | `micup.EventTypingStart` | `micup.TypingStart` | yes |
| `INVITE_CREATE` | `micup.EventInviteCreate` | `micup.Invite` | yes |
| `INVITE_DELETE` | `micup.EventInviteDelete` | `micup.InviteDelete` | yes |
| `WEBHOOKS_UPDATE` | `micup.EventWebhooksUpdate` | `micup.WebhooksUpdate` | yes |
| `INTERACTION_CREATE` | `micup.EventInteractionCreate` | `micup.Interaction` | yes |
| `APPLICATION_COMMANDS_UPDATE` | `micup.EventApplicationCommandsUpdate` | `micup.ApplicationCommandsUpdate` | yes |
| `VOICE_STATE_UPDATE` | `micup.EventVoiceStateUpdate` | `micup.VoiceState` | yes |
| `VOICE_SERVER_UPDATE` | `micup.EventVoiceServerUpdate` | `micup.VoiceServerUpdate` | yes |
<!-- events:end -->

## How payloads are kept accurate

Payload types live in `src/events`, one file per domain (`guild.go`, `member.go`, `message.go`, …). `registry.go` maps each event name to its type. Tests enforce that:

- every name in `gateway.KnownEvents()` has a registered type, and every registered name is known;
- every event has an anonymized payload in `src/events/testdata` or is listed as uncaptured in `events_test.go`;
- every captured payload decodes without error, and key fields are checked per event;
- this page lists every event with its current type.

The session cache and member inventories decode with the same types through `events.Parse`, which keeps the fields it could read when Micup sends a value of an unexpected type, so one changed field does not drop an event.

To add an event: add the name to `src/gateway/names.go` and `knownEvents`, add the payload type to the right file in `src/events`, register it in `registry.go`, add an alias to the root `events.go`, add a captured payload to `src/events/testdata` (named `EVENT_NAME.N.json`, with IDs, names, tokens and locations replaced), then run `go test -run 'TestPublicAPIIsStable|TestEventDocsAreCurrent' -update-api -update-docs .`.
