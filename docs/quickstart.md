# Installation and first request

## Requirements

Use Go 1.26.8 or newer. A bot token is required. The voice package also requires an FFmpeg executable built with `libopus` support.

```sh
go get github.com/yestefir/micup.go@v0.4.0
```

## One import

Everything public is in the root package `github.com/yestefir/micup.go` (package name `micup`). Voice publishing is the only other public package, `github.com/yestefir/micup.go/voice`, kept separate so bots without voice do not build LiveKit and WebRTC. Code under `src/` is the implementation; its paths may change between releases, so do not import it directly.

## Create a bot

Read the bot token from the environment. Do not put it in source code, command-line arguments, logs, or a checked-in environment file.

```go
import micup "github.com/yestefir/micup.go"

b, err := micup.New(ctx, micup.Config{
	Token: os.Getenv("MICUP_BOT_TOKEN"),
})
if err != nil {
	return err
}
```

`micup.New` reads Micup's `/.well-known/fluxer` document, checks that the advertised API and Gateway stay on the configured origin, builds a `*micup.RESTClient`, validates the bot identity, and reads the application's message-content grant when available. The default origin is `https://micup.gg` (`micup.DefaultOrigin`). Tests and private deployments can provide `Config.Origin` and an `http.Client`.

The result exposes:

| Field or method | Meaning |
|---|---|
| `b.REST` | The `*micup.RESTClient` for channels, messages and interactions |
| `b.ID` | The bot user ID |
| `b.MessageContentEnabled` | The message-content grant, or `nil` when unknown |
| `b.Endpoints()` | The discovered API and Gateway URLs |
| `b.NewSession`, `b.NewShardManager`, `b.Gateway` | Gateway entry points, see [the Gateway guide](gateway.md) |
| `b.SessionConfig()` | The `micup.SessionConfig` passed to `micup.NewSession` and `micup.NewShardManager`, for custom setups |

The HTTP client rejects redirects so a bot token is not forwarded to another origin. REST calls use the bot authorization scheme, bound response sizes, and honor rate-limit delays with context cancellation.

## REST without the Gateway

`micup.NewREST(micup.RESTConfig{BaseURL: apiURL, Token: token})` creates a REST client directly. `BaseURL` is the API root without the version; the client appends `/v1`. `Client.Do(ctx, method, path, body, dest)` sends any API-relative request and returns `*micup.APIError` for non-2xx responses.

## Shards

`Config.ShardID` defaults to zero and `Config.ShardCount` to one; they apply to `b.NewSession` and `b.Gateway`. `micup.ShardForGuild(guildID, shardCount)` tells which shard owns a guild. `b.NewShardManager` runs several shards at once. The library does not infer a shard count or move guilds between processes.
