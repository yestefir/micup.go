# Architecture

The root `micup` package owns Micup REST discovery and Gateway transport. The `voice` package owns the downstream LiveKit room and PCM-to-Opus encoder. Applications connect the two by handling Micup voice events, pairing the short-lived grant, and passing it to `voice.LiveKit.Connect`.

The module has no dependency on a bot framework, command router, database, NATS, media search provider, or queue. Application-specific policy stays in the caller. REST payloads remain available through raw JSON so Micup fields can be adopted without waiting for a library release.

Both Gateway entry points share one connection loop. It handles HELLO, IDENTIFY, RESUME, heartbeat ACKs, heartbeat requests, dispatch events, reconnect requests, invalid-session responses and close codes. Shard IDs are explicit.

`Client.Gateway` runs that loop once. `Session` supervises it: it reconnects with bounded backoff, resumes the last session, stops only on fatal close codes, and keeps a `State` cache of guild channels and voice states. `Session` also owns the voice join handshake (opcode 4 paired with `VOICE_SERVER_UPDATE`), the current grant per guild, and the voice-lost signal. Applications decide what a lost grant means for their own media connections.

`ShardManager` scales the same `Session` horizontally. It owns one `Session` per local shard, paces IDENTIFY through a shared limiter, routes guild-scoped calls with `ShardForGuild`, and stops all shards together on a fatal close. Shards never share connection or cache state, so a process can run any subset of the total shard count.

Inside the root package, files follow responsibility: `client.go` (setup and discovery), `rest.go`, `messages.go`, `interactions.go`, `events.go`, `gateway.go`, `session.go`, `session_voice.go`, `shard.go`, `state.go` and `types.go`. A new REST resource gets its own file next to `messages.go`; a new Gateway behavior belongs in `session.go` or a `session_*.go` file.

Command syntax, command routing, permissions, persistence and multi-process coordination stay in the application.

The voice package is optional for applications that only need REST or Gateway events. Import it only when publishing PCM to Micup's LiveKit voice service.
