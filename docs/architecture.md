# Architecture

## Public facades

Applications import `github.com/yestefir/micup.go` (package `micup`) and, for voice, `github.com/yestefir/micup.go/voice`. These two facades contain only type aliases, constants, variables and one-line forwarding functions that point into `src/`. `TestFacadesOnlyForward` in `src/internal/layers` fails if a facade declares a non-alias type or a function with logic, so behavior always lives in `src/` and the public names stay stable while `src/` is reorganized.

When a `src/` package gains a public name that applications need, add an alias or forwarding function to the facade in the same change.

## Domain packages

The implementation lives under `src/`. Each directory is one domain with one job, and dependencies point downward:

```text
layer 4   bot
layer 3   shard
layer 2   session
layer 1   state      rest       discovery
layer 0   gateway    model      internal/*

voice     independent; imports no other micup.go package
```

A package imports only packages in lower layers.

| Package | Owns | May import |
|---|---|---|
| `model` | Plain payload types (users, channels, messages, embeds, interactions, voice) | nothing |
| `discovery` | `/.well-known/fluxer`, origin and endpoint validation | `internal/httpx` |
| `rest` | Authenticated HTTP, rate limits, channel/message/interaction/user resources | `model`, `internal/httpx`, `internal/version` |
| `gateway` | Protocol: opcodes, close codes, events, resume state, one connection loop | `internal/version` |
| `state` | Guild, channel and voice-state cache built from Gateway events | `gateway`, `model` |
| `session` | One supervised shard: reconnect, backoff, resume, cache, voice join handshake | `gateway`, `model`, `state` |
| `shard` | Guild→shard mapping, multi-shard `Manager`, IDENTIFY pacing | `gateway`, `model`, `session` |
| `bot` | Composition root: discovery, identity, and wiring of the packages above | all of the above |
| `voice` | LiveKit room and PCM→Opus publishing through FFmpeg | nothing internal |
| `internal/*` | Version string, redirect-safe HTTP client, shared test fixture, layer test | — |

`src/internal/layers` contains a test that parses every package's imports and fails when a package imports something outside its row. Add a row there when you add a package.

## Flow

`micup.New` (`bot.New`) validates the config, discovers endpoints, builds a `rest.Client`, and reads the bot identity. `bot.SessionConfig()` turns that into a `session.Config` (endpoint, token, bot ID, shard pair, channel lookup). `session.New` and `shard.NewManager` take that config, so every lower package can be used and tested without `bot`.

`gateway.Run` holds one WebSocket: HELLO, IDENTIFY or RESUME, heartbeats and ACKs, dispatch, reconnect requests, invalid sessions and close codes. It knows nothing about reconnecting or caching. `session.Session` wraps it in a backoff loop, owns a `gateway.Resume`, applies each event to its `state.Cache`, then calls the application handler. `shard.Manager` owns one `Session` per local shard, passes a shared IDENTIFY limiter as `BeforeIdentify`, routes guild-scoped calls with `shard.ForGuild`, and stops every shard together on a fatal close.

`session` depends on `rest` only through the `session.ChannelLookup` interface, so the Gateway side never imports the HTTP side.

## Boundaries

The module has no dependency on a bot framework, command router, database, NATS, media search provider, or queue. Command syntax, routing, permissions, persistence and multi-process coordination stay in the application. Payloads remain available as raw JSON (`Event.Data`) so new Micup fields can be adopted without a library release.

Applications connect Gateway and voice by handling voice events, pairing the short-lived grant, and passing it to `voice.LiveKit.Connect`. The voice package is optional.

## Adding code

- A new REST resource gets its own file in `rest/` next to `messages.go`, with request/response types in `model/` when they are shared.
- A new Gateway behavior that needs connection state belongs in `session/`; protocol-only additions (opcodes, close codes, event names) belong in `gateway/`.
- A new domain gets a new directory under `src/` and a row in the layer test.
