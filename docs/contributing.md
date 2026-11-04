# Contributing

Keep Go files free of comment lines. Put setup steps, API contracts, examples, security notes, and design explanations in Markdown under `docs/` and link them from the README.

Format and verify changes from the repository root:

```sh
gofmt -w client.go client_test.go gateway.go types.go voice/voice.go
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Tests use local HTTP servers and WebSocket fixtures. Live Micup credentials and a public service are not required for the normal test suite. Voice publishing integration tests can use a local LiveKit server and FFmpeg, but must not print credentials or grant tokens.
