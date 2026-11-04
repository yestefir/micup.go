# micup.go

`micup.go` is a standalone Go client for Micup bots. It discovers the public REST and Gateway endpoints, authenticates bot requests, delivers native Gateway events, sends messages with embeds, responds privately to interactions, and can publish stereo PCM to a LiveKit voice room.

The module keeps platform transport separate from application behavior. It does not implement music search, queues, command syntax, persistence, or worker coordination.

## Requirements

- Go 1.26.8 or newer
- A Micup bot token
- FFmpeg with `libopus` enabled when publishing voice audio

## Install

```sh
go get github.com/yestefir/micup.go@v0.2.1
```

The public repository is [yestefir/micup.go](https://github.com/yestefir/micup.go). The Go module path matches that repository.

## Quick start

```go
client, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	log.Fatal(err)
}

err = client.Gateway(ctx, func(ctx context.Context, event micup.Event) error {
	if event.Name != "MESSAGE_CREATE" {
		return nil
	}
	var message micup.MessageCreate
	if err := json.Unmarshal(event.Data, &message); err != nil {
		return err
	}
	if message.Author.Bot || message.Content != "!ping" {
		return nil
	}
	_, err := client.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
})
```

See [the guides](docs/README.md) for setup, Gateway events, voice publishing, security boundaries, and examples.

## License

This project is free to use under the [MIT License](LICENSE), including in commercial projects. Keep the copyright and license notice with copies or substantial portions of the software.

## Repository layout

```text
client.go          Discovery, bot authentication, and REST requests
gateway.go         Native Gateway connection and event delivery
types.go           Micup REST and Gateway payload types
voice/             LiveKit room and PCM-to-Opus publishing
docs/              User, API, architecture, and contributor guides
```

Go source files intentionally contain no comments. The documentation site is the source of user and API guidance.

## Current scope

The first release target is the public REST/Gateway transport and the LiveKit PCM publishing helper. The Gateway runner holds one connection and returns on a disconnect; reconnect and resume policy belongs to the calling application. Identify uses the configured shard ID and count. It does not discover shard counts or redistribute shards.

The current public release is `v0.2.1`. Release history is in [CHANGELOG.md](CHANGELOG.md).
