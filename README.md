# micup.go

`micup.go` is a standalone Go client for Micup bots, following the Fluxer API that Micup runs on. It covers:

- **Gateway:** sessions that resume, replay unprocessed events and re-identify on the same socket; a shard manager for many shards in one process; optional zstd-stream compression; ignored events and presence; command pacing below Micup's limits.
- **Gateway commands:** member requests, guild and channel member counts, Lazy Request subscriptions and voice state updates; every Micup dispatch event, including threads, calls, emojis, stickers and member lists, has a typed payload.
- **State:** guilds with settings, emojis and stickers, channels and threads with permission overwrites, roles, member roles and voice states, with permission and role hierarchy checks.
- **REST:** messages (send, edit, forward, crosspost, delete, bulk delete, list, search), buttons, select menus, modals and commands, reactions, pins, typing, DMs, files (multipart and presigned uploads), channels, threads and forums, roles, guild settings, members, invites, emojis, stickers, audit logs, bans, moderation and webhooks, with paging iterators, Micup's error codes and rate limit rules.
- **Helpers:** mentions, timestamps, markdown escaping, an embed builder, media URLs, snowflake times and permission names.
- **Voice:** the join handshake with fast refusal and permission checks, grant rotation, and Opus publishing to LiveKit.

The module keeps platform transport separate from application behavior. It does not implement music search, queues, command syntax, persistence, or worker coordination.

## Requirements

- Go 1.26.8 or newer
- A Micup bot token
- FFmpeg with `libopus` enabled when publishing voice audio

## Install

```sh
go get github.com/yestefir/micup.go@v0.8.0
```

## Quick start

```go
import micup "github.com/yestefir/micup.go"

b, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	log.Fatal(err)
}

mux := micup.NewEventMux()
mux.On(micup.EventMessageCreate, micup.Listen(func(ctx context.Context, message micup.MessageCreate) error {
	if message.Author.Bot || message.Content != "!ping" {
		return nil
	}
	_, err := b.REST.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
}))

s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
if err != nil {
	log.Fatal(err)
}
err = s.Run(ctx)
```

Every event has a typed payload; see [Events and typed payloads](docs/events.md). Slash and context menu commands are registered over REST; see [Commands](docs/commands.md).

Channels, roles, threads, invites, emojis and search are described in [Channels, roles, threads and invites](docs/channels.md), and mentions, embeds and media URLs in [Formatting](docs/formatting.md).

Bots in many guilds use `b.NewShardManager(micup.ShardOptions{ShardCount: 4, Session: options})` instead of `b.NewSession`. See [Shards](docs/gateway.md#shard-manager).

## Public API

For a complete program with token validation and Ctrl+C shutdown, see the [runnable ping bot](examples/README.md).

Applications import only two paths:

| Import | Contains |
|---|---|
| `github.com/yestefir/micup.go` | Everything: bot setup, REST, Gateway events, sessions, shards, state cache, payload types, errors |
| `github.com/yestefir/micup.go/voice` | LiveKit room and PCM→Opus publishing (separate so bots without voice skip LiveKit and WebRTC) |

Both are thin facades over `src/`. Code under `src/` can be moved or split without changing these import paths or names; a test fails if a facade gains logic of its own.

See [the guides](docs/README.md) for setup, Gateway events, voice publishing, security boundaries, and the [v0.3 → v0.4 migration](docs/migration.md).

## Repository layout

The implementation lives under `src/`, one package per domain:

```text
micup.go       Public facade: package micup
events.go      Public facade: event names, payload types, event mux
voice/         Public facade: package voice
src/
  bot/         Entry point: discovery, bot identity, wiring for sessions and shards
  discovery/   /.well-known endpoint discovery and origin checks
  rest/        Authenticated REST client: channels, messages, interactions, users
  gateway/     Opcodes, close codes, event names, single-connection Gateway loop
  events/      Payload type per Gateway event, registry, typed decoding, event mux
  state/       Guild channel and voice state cache
  session/     Reconnecting, resuming session for one shard; voice join handshake
  shard/       Guild-to-shard mapping and multi-shard Manager
  model/       Domain types shared by REST and Gateway
  voice/       LiveKit room and PCM-to-Opus publishing
  internal/    Version, HTTP safety, test fixtures, layer and facade rules
docs/          User, API, architecture, and contributor guides
```

Package dependencies point one way and are enforced by a test; see [Architecture](docs/architecture.md). Go source files intentionally contain no comments.

Run `make check` before sending a change. CI runs the same checks on Linux, macOS and Windows.

## License

This project is free to use under the [MIT License](LICENSE), including in commercial projects. Keep the copyright and license notice with copies or substantial portions of the software.

The current release is `v0.8.0`. Release history is in [CHANGELOG.md](CHANGELOG.md).

`SessionOptions.TrackMembers` / `Session.MemberInventories()`: [maintained member inventories](docs/guilds.md#maintained-member-inventories).
