# Changelog

Notable changes to micup.go are recorded here.

## [0.4.0] - Unreleased

### Changed

- The implementation moved into domain packages under `src/`: `bot`, `discovery`, `rest`, `gateway`, `state`, `session`, `shard`, `model` and `voice`. Applications keep importing `github.com/yestefir/micup.go` and `github.com/yestefir/micup.go/voice`, which are now thin facades over `src/`.
- Breaking: `micup.Client` is `micup.Bot`; `BotID` is `ID`; REST methods live on `b.REST` and `Client.REST` is `b.REST.Do`. See [the migration guide](docs/migration.md).
- Breaking: `Bot.NewSession` returns an error; `ShardManagerOptions` is `ShardOptions`; `State.Apply` takes the event name and data.
- `session` reaches REST only through `micup.ChannelLookup`, and the Gateway adds its own `v` and `encoding` query parameters.
- The default User-Agent is `micup.go/0.4.0`.

### Added

- `micup.NewREST`, `micup.RunGateway`, `micup.NewSession`, `micup.NewShardManager`, `micup.NewState`, `Bot.SessionConfig`, `Bot.Endpoints`, `RESTClient.CurrentUser`, `RESTClient.CurrentApplication`, `CloseError.ResetsSession`, `State.Remember`.
- Layer and facade tests in `src/internal/layers`, and a shared Gateway test fixture in `src/internal/fixture`.
- Tests for discovery, the state cache, Gateway option validation, bot wiring and the facades.

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
