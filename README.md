# micup.go

`micup.go` is a standalone Go client for Micup bots. It discovers the public REST and Gateway endpoints, authenticates bot requests, keeps a Gateway session connected and resumed, caches guild channels and voice states, joins voice channels, sends messages with embeds, responds privately to interactions, and can publish stereo PCM to a LiveKit voice room.

The module keeps platform transport separate from application behavior. It does not implement music search, queues, command syntax, persistence, or worker coordination.

## Requirements

- Go 1.26.8 or newer
- A Micup bot token
- FFmpeg with `libopus` enabled when publishing voice audio

## Install

```sh
go get github.com/yestefir/micup.go@v0.3.0
```

The public repository is [yestefir/micup.go](https://github.com/yestefir/micup.go). The Go module path matches that repository.

## Quick start

```go
client, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	log.Fatal(err)
}

session := client.NewSession(micup.SessionOptions{Handler: func(ctx context.Context, event micup.Event) error {
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
}})
err = session.Run(ctx)
```

Bots in many guilds can replace `client.NewSession` with `client.NewShardManager(micup.ShardManagerOptions{ShardCount: 4, Session: options})`. See [Shards](docs/gateway.md#shard-manager).

See [the guides](docs/README.md) for setup, Gateway events, voice publishing, security boundaries, and examples.

## License

This project is free to use under the [MIT License](LICENSE), including in commercial projects. Keep the copyright and license notice with copies or substantial portions of the software.

## Repository layout

```text
client.go          Config, endpoint discovery, bot authentication
rest.go            Authenticated REST requests and rate-limit retries
messages.go        Channel lookup and channel messages
interactions.go    Interaction responses, deferrals, follow-ups
events.go          Event type and Gateway event name constants
gateway.go         Opcodes, close codes, single-connection Gateway loop
session.go         Reconnecting, resuming session for one shard
session_voice.go   Voice join handshake, grants, voice-lost signal
shard.go           Shard mapping and multi-shard ShardManager
state.go           Guild channel and voice state cache
types.go           Micup REST and Gateway payload types
version.go         Module version and default User-Agent
voice/             LiveKit room and PCM-to-Opus publishing
docs/              User, API, architecture, and contributor guides
```

Run `make check` before sending a change. CI runs the same checks on Linux, macOS and Windows.

Go source files intentionally contain no comments. The documentation site is the source of user and API guidance.

## Current scope

The module covers the public REST/Gateway transport, a supervised Gateway session with resume, a guild/voice state cache, the voice join handshake, the LiveKit PCM publishing helper, and a `ShardManager` that runs many shards in one process with paced IDENTIFY. Shard counts are configured by the application; the module does not discover them or move shards between processes.

The current release is `v0.3.0`. Release history is in [CHANGELOG.md](CHANGELOG.md).
