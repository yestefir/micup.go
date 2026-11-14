# Known issues and investigation candidates

## Confirmed unresolved validation gaps

| Priority | Issue | Impact and evidence | Next action |
|---|---|---|---|
| P2 | Optional voice integration checks skipped locally | `TestPCMTrackEncodesOpus` skips because FFmpeg is absent; `TestLocalLiveKitReceivesOpusAudio` skips because `LIVEKIT_TEST_BINARY` is unset. The remaining tests pass, but real audio encoding/delivery is not verified here. | Use local test dependencies when available; rerun those tests without production services. |
| P3 | Full formatting check fails on 124 existing CRLF files | Clean initial Git tree; `core.autocrlf=true`; byte inspection finds 124 CRLF Go files; `gofmt -d src/rest/client.go` shows line-ending-only changes. Changed Go files are gofmt clean. | Consider a focused LF attributes/checkout policy change, separately from behavior fixes. Do not claim `make check` passed. |

## Unverified candidates

- `src/rest/webhooks.go`: `ExecuteWebhook` omits `anonymous` unlike the token-based methods. Creation variants also use different byte/rune length validation. These inconsistencies are visible in source; practical failures and intended platform requirements remain to be established.
- `src/rest/errors.go`: float-to-duration conversion lacks explicit overflow/nonfinite checks. Reproduce the effect on retry policy with controlled headers/bodies before classifying severity.

## Scope limitations

- New iterator progress guards detect an empty or unchanged next cursor. They do not deduplicate previously yielded items or detect arbitrary multi-page cursor cycles. A repeated page may be yielded once before termination, matching the existing member/thread iterator convention.
- Discovery sampled each domain, documentation, tests and critical REST/session/Gateway paths. This cycle was not an exhaustive audit of every source line or dependency advisory.
- Runtime latency, CPU, allocation rate, memory usage and production throughput have no baseline. No performance improvement is claimed.

## Resolved in cycle 2026-11-14_03-47

- P1: Shared mutable message/audit iterator query produced confirmed data races under concurrent use and skipped earlier pages on subsequent traversals. Fixed with a per-traversal query copy.
- P1: Message, audit-log and guild iteration could keep fetching a repeated full page. Fixed with cursor progress checks, verified against bounded local server fixtures.
