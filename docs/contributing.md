# Contributing

Keep Go files free of comment lines. Put setup steps, API contracts, examples, security notes, and design explanations in Markdown under `docs/` and link them from the README.

Run `make check` from the repository root. It runs the same steps as CI (gofmt check, vet, build and the race-enabled test suite). The individual commands are:

```sh
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Tests live next to the package they cover. Session and shard tests share the fake Micup server in `src/internal/fixture`. `src/internal/layers` fails when a package imports outside its allowed layer; see [Architecture](architecture.md). Tests use local HTTP servers and WebSocket fixtures. Live Micup credentials and a public service are not required for the normal test suite. Voice publishing integration tests can use a local LiveKit server and FFmpeg, but must not print credentials or grant tokens.
