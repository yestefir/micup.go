# Architecture

The root `micup` package owns Micup REST discovery and Gateway transport. The `voice` package owns the downstream LiveKit room and PCM-to-Opus encoder. Applications connect the two by handling Micup voice events, pairing the short-lived grant, and passing it to `voice.LiveKit.Connect`.

The module has no dependency on a bot framework, command router, database, NATS, media search provider, or queue. Application-specific policy stays in the caller. REST payloads remain available through raw JSON so Micup fields can be adopted without waiting for a library release.

The Gateway runner uses one live connection per call. Shard IDs are explicit. It handles HELLO, IDENTIFY, heartbeat ACKs, heartbeat requests, dispatch events, reconnect requests, and invalid-session responses. It does not automatically reconnect or resume a session. A caller should supervise each shard and apply bounded backoff.

The voice package is optional for applications that only need REST or Gateway events. Import it only when publishing PCM to Micup's LiveKit voice service.
