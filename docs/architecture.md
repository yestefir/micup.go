# Architecture

## Public facades

Applications import `github.com/yestefir/micup.go` (package `micup`) and, for voice, `github.com/yestefir/micup.go/voice`. These two facades contain only type aliases, constants, variables and one-line forwarding functions that point into `src/`. `TestFacadesOnlyForward` in `src/internal/layers` fails if a facade declares a non-alias type or a function with logic, so behavior always lives in `src/` and the public names stay stable while `src/` is reorganized.

When a `src/` package gains a public name that applications need, add an alias or forwarding function to the facade in the same change.

## Domain packages

The implementation lives under `src/`. Each directory is one domain with one job, and dependencies point downward:

```text
layer 5   bot
layer 4   shard
layer 3   session
layer 2   state
layer 1   events     rest       discovery
layer 0   gateway    model      internal/*

voice     independent; imports no other micup.go package
```

A package imports only packages in lower layers.

| Package | Owns | May import |
|---|---|---|
| `model` | Domain types shared by REST and Gateway (users, guilds, channels, roles, messages, embeds, buttons, interactions, invites, voice) | nothing |
| `discovery` | `/.well-known/fluxer`, origin and endpoint validation | `internal/httpx` |
| `rest` | Authenticated HTTP, rate limits, channel/message/interaction/user resources | `model`, `internal/httpx`, `internal/version` |
| `gateway` | Protocol: opcodes, close codes, event names, resume state, one connection loop | `internal/version`, `model` |
| `events` | Gateway payload types per event, the name→type registry, typed decoding and the event mux | `gateway`, `model` |
| `state` | Guild, channel and voice-state cache built from Gateway events | `events`, `gateway`, `model` |
| `session` | One supervised shard: reconnect, backoff, resume, cache, voice join handshake | `events`, `gateway`, `model`, `state` |
| `shard` | Guild→shard mapping, multi-shard `Manager`, IDENTIFY pacing | `gateway`, `model`, `session` |
| `bot` | Composition root: discovery, identity, and wiring of the packages above | all of the above |
| `voice` | LiveKit room and PCM→Opus publishing through FFmpeg | nothing internal |
| `internal/*` | Version string, redirect-safe HTTP client, shared test fixture, layer test, payload anonymizer | — |

`src/internal/layers` contains a test that parses every package's imports and fails when a package imports something outside its row. Add a row there when you add a package.

## Flow

`micup.New` (`bot.New`) validates the config, discovers endpoints, builds a `rest.Client`, and reads the bot identity. `bot.SessionConfig()` turns that into a `session.Config` (endpoint, token, bot ID, shard pair, channel lookup). `session.New` and `shard.NewManager` take that config, so every lower package can be used and tested without `bot`.

`gateway.Run` holds one WebSocket: HELLO, IDENTIFY or RESUME, heartbeats and ACKs, dispatch, reconnect requests, invalid sessions and close codes. It knows nothing about reconnecting or caching. `session.Session` wraps it in a backoff loop, owns a `gateway.Resume`, applies each event to its `state.Cache`, then calls the application handler. `shard.Manager` owns one `Session` per local shard, passes a shared IDENTIFY limiter as `BeforeIdentify`, routes guild-scoped calls with `shard.ForGuild`, and stops every shard together on a fatal close.

`session` depends on `rest` only through the `session.ChannelLookup` and `session.MemberLookup` interfaces, so the Gateway side never imports the HTTP side.

## Boundaries

The module has no dependency on a bot framework, command router, database, NATS, media search provider, or queue. Command syntax, routing, permissions, persistence and multi-process coordination stay in the application. Payloads remain available as raw JSON (`Event.Data`) so new Micup fields can be adopted without a library release.

Applications connect Gateway and voice by handling voice events, pairing the short-lived grant, and passing it to `voice.LiveKit.Connect`. The voice package is optional.

## Adding code

- A new REST resource gets its own file in `rest/` (`messages.go`, `reactions.go`, `pins.go`, `attachments.go`, `members.go`, `guilds.go`, `channels.go`, `roles.go`, `threads.go`, `invites.go`, `expressions.go`, `webhooks.go`; `iterate.go` holds the paging iterators) and goes through `Client.call`, which applies rate limits, retries, audit reasons and multipart bodies. Shared request and response types live in `model/`.
- Gateway protocol pieces live in `gateway/`: `run.go` (connection loop), `conn.go` (send pacing and size checks), `compress.go` (zstd-stream), `limiter.go`, `options.go`, `event.go`, `names.go` (event names and `KnownEvents`), `errors.go`. Session features live in `session/`: `voice.go` (join handshake, grants, refusals), `members.go` (opcode 8), `requests.go` (opcodes 14, 15 and 16), `permissions.go`. Edit structs use `model.Optional` so a field can be left out, set or cleared.
- A new Gateway behavior that needs connection state belongs in `session/`; protocol-only additions (opcodes, close codes, event names) belong in `gateway/`.
- A new event payload goes in `events/` beside its domain and is registered in `events/registry.go`; see [Events and typed payloads](events.md#how-payloads-are-kept-accurate). Code that reads a payload uses the `events` type, never a private struct.
- The facade's `api_test.go` records every public name in `testdata/api.golden`; refresh it with `-update-api` after adding names. Removing a name fails the test.
- A new domain gets a new directory under `src/` and a row in the layer test.

Optional member inventories live in `session/member_inventory.go`. The session owns pagination and event reconciliation; application statistics, persistence and worker coordination remain outside the module. See [member inventories](guilds.md#maintained-member-inventories).
