# Installation and first request

## Requirements

Use Go 1.26.8 or newer. A bot token is required. The voice package also requires an FFmpeg executable built with `libopus` support.

Install the module:

```sh
go get github.com/yestefir/micup.go
```

## Create a client

Read the bot token from the environment. Do not put it in source code, command-line arguments, logs, or a checked-in environment file.

```go
client, err := micup.New(ctx, micup.Config{
	Token: os.Getenv("MICUP_BOT_TOKEN"),
})
if err != nil {
	return err
}
```

The module is published from [github.com/yestefir/micup.go](https://github.com/yestefir/micup.go). `New` reads Micup's `/.well-known/fluxer` document, checks that the advertised API and Gateway stay on the configured origin, validates the bot identity, and reads the application's message-content grant when available. The default origin is `https://micup.gg`. Tests and private deployments can provide `Config.Origin` and an `http.Client`.

The HTTP client rejects redirects so a bot token is not forwarded to another origin. REST calls use the bot authorization scheme, bound response sizes, and honor rate-limit delays with context cancellation.

## Shards

`ShardID` defaults to zero and `ShardCount` defaults to one. Configure both when connecting a shard. The library does not infer a shard count or move guilds between running connections.
