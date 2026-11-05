# Installation and first request

## Requirements

Use Go 1.26.8 or newer. A bot token is required. The voice package also requires an FFmpeg executable built with `libopus` support.

```sh
go get github.com/yestefir/micup.go@v0.4.0
```

## Create a bot

Read the bot token from the environment. Do not put it in source code, command-line arguments, logs, or a checked-in environment file.

```go
import "github.com/yestefir/micup.go/src/bot"

b, err := bot.New(ctx, bot.Config{
	Token: os.Getenv("MICUP_BOT_TOKEN"),
})
if err != nil {
	return err
}
```

`bot.New` reads Micup's `/.well-known/fluxer` document (package `discovery`), checks that the advertised API and Gateway stay on the configured origin, builds a `rest.Client`, validates the bot identity, and reads the application's message-content grant when available. The default origin is `https://micup.gg` (`discovery.DefaultOrigin`). Tests and private deployments can provide `Config.Origin` and an `http.Client`.

The result exposes:

| Field or method | Meaning |
|---|---|
| `b.REST` | The `*rest.Client` for channels, messages and interactions |
| `b.ID` | The bot user ID |
| `b.MessageContentEnabled` | The message-content grant, or `nil` when unknown |
| `b.Endpoints()` | The discovered API and Gateway URLs |
| `b.NewSession`, `b.NewShardManager`, `b.Gateway` | Gateway entry points, see [the Gateway guide](gateway.md) |
| `b.SessionConfig()` | The wiring passed to `session.New` and `shard.NewManager`, for custom setups |

The HTTP client rejects redirects so a bot token is not forwarded to another origin. REST calls use the bot authorization scheme, bound response sizes, and honor rate-limit delays with context cancellation.

## REST without the Gateway

`rest.New(rest.Config{BaseURL: apiURL, Token: token})` creates a REST client directly. `BaseURL` is the API root without the version; the client appends `/v1`. `Client.Do(ctx, method, path, body, dest)` sends any API-relative request and returns `*rest.APIError` for non-2xx responses.

## Shards

`Config.ShardID` defaults to zero and `Config.ShardCount` to one; they apply to `b.NewSession` and `b.Gateway`. `shard.ForGuild(guildID, shardCount)` tells which shard owns a guild. `b.NewShardManager` runs several shards at once. The library does not infer a shard count or move guilds between processes.
