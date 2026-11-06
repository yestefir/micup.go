# Changelog

Notable changes to micup.go are recorded here.

## [Unreleased]

### Added

- `src/events`: a payload type for every known Gateway event, grouped by domain, with a name→type registry. `micup.EventAs[T]` decodes an event and rejects the wrong type; `micup.DecodeEvent` returns the payload as `any` and raw JSON for unknown names.
- `micup.NewEventMux` with `On(name, micup.Listen(fn))`, `HandleRaw` and `HandleAll`. A listener whose type does not match the event panics at registration.
- New payload types: `Ready`, `GuildCreate`, `GuildUpdate`, `GuildDelete`, `RoleUpdateBulk`, `MembersChunk`, `GuildBan`, `ChannelUpdateBulk`, `ChannelPinsUpdate`, `ReactionRemoveAll`, `ReactionRemoveEmoji`, `PresenceUpdate`, `InviteDelete`, `WebhooksUpdate`, `RateLimited`, `SessionsReplace`, `AuditLogEntryCreate`.
- Event names seen on Micup: `EventGuildAuditLogEntryCreate`, `EventSessionsReplace` and `EventGatewayRolloutUpdate` (`GatewayRollout`, also `Ready.GatewayRollout`). `micup.KnownEvents()` lists every name.
- Model fields Micup returns: guild settings and counts, member avatar, banner and flags, role colors, emoji and `managed`, channel `last_pin_timestamp`, user discriminator and flags, voice state mute, deaf, stream and member. New `Invite` and `RoleColors` types. `Bot.ApplicationID`.
- [Events and typed payloads](docs/events.md), with an event table generated from the registry.
- Application commands: `GlobalCommands`, `GuildCommands`, `Create…`, `Edit…`, `Delete…` and `Set…Commands` for global and guild scopes, with `ApplicationCommand.Validate` checking Micup's rules before a request is sent. See [Slash and context menu commands](docs/commands.md).
- `Interaction.CommandData` with `Path`, typed option readers, resolved users, members, roles, channels and attachments, `Focused`, `TargetUser` and `TargetMessage`. `IsCommand`, `IsAutocomplete` and `RespondAutocomplete`.
- `Interaction.Channel`, `Guild`, `Context` and `AttachmentSizeLimit`; `Member.Permissions` and `Channel.Permissions` as resolved by Micup. `EventApplicationCommandsUpdate` and `ApplicationCommandsUpdate`.

### Fixed

- Concurrent `RequestGuildMembers` calls on one session could time out: Micup silently drops an opcode 8 request while two others are being answered, without a `RATE_LIMITED` event. A session now keeps at most two member requests in flight and queues the rest. Measured on Micup: 13 of 15 simultaneous requests were dropped before, none after.

### Changed

- Payload types were checked against anonymized responses and events captured from Micup; those captures are part of the test suite.
- The session cache and member inventories decode payloads with the shared types and keep the readable fields when Micup sends an unexpected value type.

## [0.6.0] - 2026-11-05

### Added

- `SetStatus`, `SetCustomStatus`, `ClearCustomStatus` and `Presence` on `Session` and `ShardManager`. Each helper changes one part of the presence and keeps the rest.
- `Session.Latency()` returns the Gateway heartbeat round trip of the live connection, for ping commands and health checks.
- `CustomStatus.Until(t)` sets the expiry. `Presence.Validate` and `CustomStatus.Validate` check Micup's limits, exported as `MaxCustomStatusText` and `MaxCustomStatusEmoji`.

### Fixed

- A custom status could not be removed. A nil `CustomStatus` left the field out, and Micup keeps the previous custom status when the field is missing. `Presence` now sends `custom_status: null`, so `UpdatePresence` with a nil `CustomStatus` clears it.
- Micup drops an invalid presence update silently. Presence changes and `SessionOptions.Presence` now return an error for text over 128 characters, an empty custom status, a malformed or past expiry, or an unknown status, and send nothing.

## [0.5.0] - 2026-11-05

### Added

- Opt-in `SessionOptions.TrackMembers` and `Session.MemberInventories()` with per-guild user IDs, bot flags, availability and completeness. The session owns REST pagination, concurrent member-event reconciliation, READY/guild-sync resets and bounded retries.
- `MemberLookup` and `SessionConfig.Members` for standalone sessions; `Bot.SessionConfig()` supplies the REST client automatically. Member requests stop with `Session.Run`.

## [0.4.0] - 2026-11-05

### Fixed

- The Gateway acknowledged an event's sequence before its handler finished, so an event whose handler failed was not replayed after RESUME. Sequences are now acknowledged only after the handler returns `nil`.
- An invalid session closed the socket and reconnected with backoff. The session now identifies again on the same socket, as Micup allows.
- A region change grant was reported as voice loss. `SessionOptions.OnVoiceGrant` now receives unrequested grants.
- `JoinVoice` waited for its context when Micup refused the join with opcode 12. It now returns `*VoiceRefusedError` at once.
- `GUILD_DELETE` with `unavailable: true` cleared the cached guild. A temporary outage now keeps it.
- A new READY dropped stored voice grants without calling `OnVoiceLost`.
- Removed close codes 4013 and 4014, which Micup does not use, from the fatal set, and removed the unused `MessageCreate.Raw` field and `ErrSessionInvalidated`.

### Changed

- The implementation moved into domain packages under `src/`: `bot`, `discovery`, `rest`, `gateway`, `state`, `session`, `shard`, `model` and `voice`. Applications keep importing `github.com/yestefir/micup.go` and `github.com/yestefir/micup.go/voice`, which are now thin facades over `src/`.
- Breaking: `micup.Client` is `micup.Bot`; `BotID` is `ID`; REST methods live on `b.REST` and `Client.REST` is `b.REST.Do`. See [the migration guide](docs/migration.md).
- Breaking: `Bot.NewSession` returns an error; `ShardManagerOptions` is `ShardOptions`; `State.Apply` takes the event name and data.
- `session` reaches REST only through `micup.ChannelLookup`, and the Gateway adds its own `v` and `encoding` query parameters.
- The default User-Agent is `micup.go/0.4.0`.
- Breaking: `Session.Send` and `GatewayConn.Send` take a context.
- `APIError` carries Micup's `Code`, `Message`, field `Errors`, `RequestID`, `RetryAfter` and `Global`.
- REST rate limiting follows Micup's rules: global 429s pause every route, exhausted buckets delay the next request on that route, 423 is retried, and 5xx is retried only for idempotent requests.
- Breaking: `Interaction.Member` is `*micup.Member` and `Interaction.User` is `*micup.User` instead of raw JSON.
- `ExecuteWebhook` returns an error when the message has buttons.
- The docs no longer say that interactions are limited to official Micup apps; ordinary bots receive button clicks.

### Added

- `micup.WithReason` for audit log reasons, and `micup.File` for multipart uploads.
- Message REST: `Message`, `Messages`, `EditMessage`, `DeleteMessage`, `BulkDeleteMessages`, `TriggerTyping`, `CreateDM`, reactions (`AddReaction`, `RemoveReaction`, `RemoveUserReaction`, `RemoveEmojiReactions`, `RemoveAllReactions`) and pins (`PinMessage`, `UnpinMessage`, `PinnedMessages`). See [Messages](docs/messages.md).
- `SessionOptions.IgnoredEvents`, `Presence` and `Flags` for IDENTIFY; `Session.UpdatePresence` and `ShardManager.UpdatePresence` (opcode 3); `micup.Presence`, `CustomStatus` and status constants.
- `SessionOptions.Compress` and `GatewayOptions.Compress` turn on Micup's `zstd-stream` Gateway compression in both directions.
- Each Gateway connection paces commands below Micup's 600 per minute and presence updates at 5 per 20 seconds, and rejects payloads over 16,384 bytes before sending.
- The state cache keeps roles, the guild owner, member roles and full channels with permission overwrites, handles `GUILD_SYNC`, `GUILD_UPDATE`, role, member, bulk channel and member chunk events, and computes permissions with `State.Permissions` and `State.BotPermissions`. `micup.Permissions`, every permission constant, `micup.ComputePermissions` and channel type constants.
- `JoinVoice` returns `ErrVoicePermission` at once when the bot lacks `VIEW_CHANNEL` or `CONNECT`, which Micup would ignore silently.
- `Session.RequestGuildMembers` and `ShardManager.RequestGuildMembers` (opcode 8) collect member chunks and report `RATE_LIMITED` as `MembersRateLimitedError`.
- Guild REST: `Guild`, `GuildChannels`, `Roles`, `Member`, `Members`, `ModifyMember` (including voice move and disconnect), `AddMemberRole`, `RemoveMemberRole`, `KickMember`, `BanMember` and `UnbanMember`. See [Guilds](docs/guilds.md).
- Event name constants for every Micup dispatch the module handles, and typed payloads: `MessageDelete`, `MessageDeleteBulk`, `ReactionEvent`, `TypingStart`, `GuildMemberRemove`, `RoleEvent`, `RoleDelete`.
- Webhooks: `CreateWebhook`, `ChannelWebhooks`, `GuildWebhooks`, `DeleteWebhook` and `ExecuteWebhook` with per-message username and avatar.
- `voice.LiveKit.Bitrate` with `DefaultBitrate`, `MinBitrate` and `MaxBitrate`.
- Documented Micup limits the module does not work around: 2,500 guilds per shard, no peer-to-peer or encrypted-media voice for this publisher, and 600-second voice grants.
- Files: `MessageOptions.Files` sends multipart, and `UploadAttachment` runs Micup's presigned single- and multipart upload flow for large files.
- `MessageOptions` gains `Flags`, `TTS`, `StickerIDs`, `AllowedMentions` and `Attachments`; `Message` gains the full Micup message fields; `MessageCreate` embeds `Message` and adds `ChannelType`, `MentionHere` and `Member`. New message type and flag constants.
- `micup.NewREST`, `micup.RunGateway`, `micup.NewSession`, `micup.NewShardManager`, `micup.NewState`, `Bot.SessionConfig`, `Bot.Endpoints`, `RESTClient.CurrentUser`, `RESTClient.CurrentApplication`, `CloseError.ResetsSession`, `State.Remember`.
- Layer and facade tests in `src/internal/layers`, and a shared Gateway test fixture in `src/internal/fixture`.
- Tests for discovery, the state cache, Gateway option validation, bot wiring and the facades.
- Buttons: `micup.ActionRow`, `micup.Button`, `NewActionRow`, `PrimaryButton`, `SecondaryButton`, `SuccessButton`, `DangerButton`, `LinkButton`, `Button.WithEmoji`, `Button.WithDisabled`, `DisableButtons`, and style, component type and limit constants. See [Buttons](docs/components.md).
- `MessageOptions.Components`, `MessageEdit.Components` and `Message.Components`. The client checks row, button, label and custom ID rules before sending.
- Button clicks: `Interaction.IsButton`, `CustomID`, `ComponentData`, `Invoker`, the `Message` the button is on, `Locale`, `GuildLocale`, `AppPermissions`, and `InteractionType*` constants.
- `UpdateInteractionMessage` (callback 7), `DeferInteractionUpdate` (callback 6) and `EditInteractionResponse`. `RespondToInteraction` and `SendInteractionFollowup` carry buttons.

## [0.3.0] - 2026-11-05

### Added

- `Session`: a supervised Gateway connection with bounded reconnect backoff, RESUME, fatal close-code detection, and a ready flag.
- `State`: guild channel and voice state cache, plus `Session.ResolveGuild` with a cached REST fallback.
- `Session.JoinVoice`, `Session.LeaveVoice`, `Session.Voice`, and `SessionOptions.OnVoiceLost` for the voice join handshake.
- `Client.Channel`, `ShardForGuild`, Gateway opcode constants, `CloseError`, and `MessageCreate.Type` with message type constants.
- `VoiceServerUpdate.ChannelID`.
- `voice.LiveKit.TrackName`.
- `ShardManager` (`Client.NewShardManager`): runs many shards in one process with paced IDENTIFY, guild-to-shard routing for voice and channel lookup, and coordinated shutdown on a fatal close.
- `Session.ShardID`, `Event.ShardID`, `Event.Decode`, Gateway event name constants (`EventReady`, `EventMessageCreate`, ...), and `Version`.
- `Makefile` with `make check`, and a GitHub Actions CI workflow for Linux, macOS and Windows.

### Changed

- `voice` track close lets FFmpeg flush buffered audio for up to one second before stopping.
- On Windows, the FFmpeg encoder starts without a console window.
- The root package is split by responsibility (`rest.go`, `messages.go`, `interactions.go`, `events.go`, `session_voice.go`, `shard.go`, `version.go`). The public API is unchanged.
- The default User-Agent is now `micup.go/0.3.0`, derived from `Version`.
- `Client.Gateway` returns `*CloseError`, `ErrReconnectRequested`, or `ErrSessionInvalidated` for the matching disconnect causes, and times out a connection that sends no HELLO within 10 seconds.

## [0.2.1] - 2026-11-05

- Support idempotent channel message creation with a nonce in `MessageOptions`.

## [0.2.0] - 2026-11-05

- Add rich embeds to channel messages and interaction responses.
- Add ephemeral interaction responses, deferrals, and follow-up messages.
- Expose the discovered Gateway endpoint for platform adapters.

## [0.1.0] - 2026-11-05

Initial public release.

### Added

- Micup endpoint discovery and bot-authenticated REST requests.
- Native Gateway connection with explicit shard settings, heartbeats, and event callbacks.
- Channel message replies with automatic mentions disabled by default.
- LiveKit room connection and PCM-to-Opus voice publishing through FFmpeg.
- English and Turkish entry documentation, API guides, architecture, security guidance, and MIT licensing.

### Known limits

- The Gateway runner returns on disconnect; applications own reconnect and resume policy.
- Shard counts must be configured by the application.
- The voice package publishes outgoing PCM only and requires FFmpeg with `libopus`.
