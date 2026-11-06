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

## Captured Micup payloads

Payload types are tested against real Micup responses and Gateway events, anonymized and stored in `src/events/testdata` (events, `EVENT_NAME.N.json`) and `src/model/testdata/rest` (REST responses). Use a test bot in a guild that holds no real users' data, and a token kept out of the repository.

1. Record raw payloads into one directory per event name, for example with a Gateway handler:

   ```go
   handler := func(ctx context.Context, e micup.Event) error {
   	dir := filepath.Join("captures", e.Name)
   	_ = os.MkdirAll(dir, 0o700)
   	entries, _ := os.ReadDir(dir)
   	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.json", len(entries))), e.Data, 0o600)
   }
   ```

2. Anonymize them, keeping the mapping file outside the repository so later captures reuse the same fake IDs:

   ```sh
   go run ./src/internal/anonymize -in captures -out /tmp/anonymized -map ~/.config/micup/fixture-map.json -keep <bot username>
   ```

   The tool replaces snowflake IDs (also when they are object keys, as in `resolved`) with stable fake IDs in document order, user names with `userN`, image hashes, invite codes, tokens, session and connection IDs, and location fields. Permission strings and audit-log permission changes are kept. Review the output before copying it: message text, channel names and custom status text are kept as they are.

3. Copy the files into the testdata directory, remove the event from `uncaptured` in `src/events/events_test.go` if it is listed, add field checks for anything new, and refresh the generated docs with `go test -run TestEventDocsAreCurrent -update-docs .`.

The mapping file contains real IDs. Never commit it or the raw captures.
