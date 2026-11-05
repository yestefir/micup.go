# micup.go

`micup.go` is a standalone Go client for Micup bots. It discovers the public REST and Gateway endpoints, authenticates bot requests, keeps Gateway sessions connected and resumed, runs many shards in one process, caches guild channels and voice states, joins voice channels, sends messages with embeds, responds privately to interactions, and can publish stereo PCM to a LiveKit voice room.

The module keeps platform transport separate from application behavior. It does not implement music search, queues, command syntax, persistence, or worker coordination.

## Requirements

- Go 1.26.8 or newer
- A Micup bot token
- FFmpeg with `libopus` enabled when publishing voice audio

## Install

```sh
go get github.com/yestefir/micup.go@v0.4.0
```

## Quick start

```go
import (
	"github.com/yestefir/micup.go/src/bot"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/session"
)

b, err := bot.New(ctx, bot.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	log.Fatal(err)
}

s, err := b.NewSession(session.Options{Handler: func(ctx context.Context, event gateway.Event) error {
	if event.Name != gateway.EventMessageCreate {
		return nil
	}
	var message model.MessageCreate
	if err := event.Decode(&message); err != nil {
		return err
	}
	if message.Author.Bot || message.Content != "!ping" {
		return nil
	}
	_, err := b.REST.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
}})
if err != nil {
	log.Fatal(err)
}
err = s.Run(ctx)
```

Bots in many guilds use `b.NewShardManager(shard.Options{ShardCount: 4, Session: options})` instead of `b.NewSession`. See [Shards](docs/gateway.md#shard-manager).

See [the guides](docs/README.md) for setup, Gateway events, voice publishing, security boundaries, and the [v0.3 → v0.4 migration](docs/migration.md).

## Repository layout

All Go code lives under `src/`, one package per domain:

```text
src/
  bot/         Entry point: discovery, bot identity, wiring for sessions and shards
  discovery/   /.well-known endpoint discovery and origin checks
  rest/        Authenticated REST client: channels, messages, interactions, users
  gateway/     Opcodes, close codes, events, single-connection Gateway loop
  state/       Guild channel and voice state cache
  session/     Reconnecting, resuming session for one shard; voice join handshake
  shard/       Guild-to-shard mapping and multi-shard Manager
  model/       Plain Micup payload types
  voice/       LiveKit room and PCM-to-Opus publishing
  internal/    Version, HTTP safety, test fixtures, layer rules
docs/          User, API, architecture, and contributor guides
```

Package dependencies point one way and are enforced by a test; see [Architecture](docs/architecture.md). Go source files intentionally contain no comments.

Run `make check` before sending a change. CI runs the same checks on Linux, macOS and Windows.

## License

This project is free to use under the [MIT License](LICENSE), including in commercial projects. Keep the copyright and license notice with copies or substantial portions of the software.

The current release is `v0.4.0`. Release history is in [CHANGELOG.md](CHANGELOG.md).
