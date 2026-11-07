# Installation and first request

## Requirements

Use Go 1.26.8 or newer. A bot token is required. The voice package also requires an FFmpeg executable built with `libopus` support.

```sh
go get github.com/yestefir/micup.go@v0.7.0
```

## One import

For a complete bot instead of individual snippets, run the [ping example](../examples/README.md). It includes token validation, a typed message handler, and Ctrl+C shutdown.

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

`micup.NewREST(micup.RESTConfig{BaseURL: apiURL, Token: token})` creates a REST client directly. `BaseURL` is the API root without the version; the client appends `/v1`. `Client.Do(ctx, method, path, body, dest)` sends any API-relative request.

## Errors

A failed request returns `*micup.APIError`:

| Field | Meaning |
|---|---|
| `StatusCode` | HTTP status |
| `Code` | Micup's stable error code, such as `MISSING_PERMISSIONS` or `UNKNOWN_MESSAGE`; branch on this |
| `Message` | Translated text for logs; do not parse it |
| `Errors` | Field errors (`Path`, `Code`, `Message`) on `INVALID_FORM_BODY` |
| `RequestID` | The `X-Request-ID` header, for bug reports |
| `RetryAfter`, `Global` | Set on rate limits and cooldowns |

## Rate limits and retries

The REST client handles Micup's limits for you:

- A 429 waits `retry_after` (fractional seconds from the body) and retries. With `global: true` it pauses every route of that client, not only the one that hit the limit. Waits over 30 s return the error instead.
- After a response with `X-RateLimit-Remaining: 0`, the next request to the same route waits for `X-RateLimit-Reset-After`. Routes are keyed by method and path, with channel, guild and webhook IDs kept, so one channel's limit does not slow another.
- A 423 (`RESOURCE_LOCKED`) waits `Retry-After` and retries.
- 5xx responses and connection failures are retried with backoff only for `GET`, `PUT` and `DELETE`, and for message sends that carry a `Nonce`. Other `POST` and `PATCH` requests are not retried, so they are never applied twice.
- At most 5 attempts are made, and every wait stops when the context ends.

Routes that write the audit log take `micup.WithReason("text")`, which sends `X-Audit-Log-Reason` (up to 512 characters).

## Shards

`Config.ShardID` defaults to zero and `Config.ShardCount` to one; they apply to `b.NewSession` and `b.Gateway`. `micup.ShardForGuild(guildID, shardCount)` tells which shard owns a guild. `b.NewShardManager` runs several shards at once. The library does not infer a shard count or move guilds between processes.
