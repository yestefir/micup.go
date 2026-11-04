# Security and release checklist

## Credentials

- Read the bot token from a secret store or process environment.
- Never commit tokens, `.env` files, LiveKit grants, cookies, media URLs, or Authorization headers.
- Treat voice endpoint tokens as short-lived credentials. Keep them in memory only.
- Avoid logging raw Gateway packets, REST bodies, extractor URLs, or token-bearing errors.

## Network boundaries

Client discovery requires REST and Gateway endpoints to use the configured Micup origin. REST redirects are disabled. Requests accept only API-relative paths. Contexts should have deadlines around application operations; rate-limit waits stop when their context is cancelled.

## Before publishing

1. Keep the GitHub module path in `go.mod`, README install commands, and examples aligned with `github.com/yestefir/micup.go`.
2. Keep the MIT license text and copyright notice with distributed copies.
3. Search the repository and Git history for credentials and generated local configuration.
4. Run the full test suite, `go vet ./...`, and `go test -race ./...`.
5. Tag releases only after the exported API and docs agree.

The public v0.1.0 release is available from the tagged commit.
