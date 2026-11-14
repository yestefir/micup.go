# Known issues and investigation candidates

## Confirmed unresolved validation gaps

| Priority | Issue | Impact and evidence | Next action |
|---|---|---|---|
| P2 | Optional voice integration checks skipped locally | `TestPCMTrackEncodesOpus` skips because FFmpeg is absent; `TestLocalLiveKitReceivesOpusAudio` skips because `LIVEKIT_TEST_BINARY` is unset. The remaining tests pass, but real audio encoding/delivery is not verified here. | Use local test dependencies when available; rerun those tests without production services. |
| P3 | Full formatting check fails on 121 remaining existing CRLF files | `core.autocrlf=true`; initial cycle verified CRLF-only formatting differences. Formatting the two webhook files modified this cycle reduced the remaining count from 123 to 121; unrelated files are untouched. Changed Go files are gofmt clean. | Consider a focused LF attributes/checkout policy change, separately from behavior fixes. Do not claim `make check` passed. |

## Unverified candidates

- `ExecuteWebhook` sends bot Authorization while token-based lookup/edit methods omit it. The existing `TestWebhookEndpoints` explicitly requires that execution behavior. No actual authorization failure has been reproduced; retain the tested contract and reopen only with new upstream evidence. The creation-name discrepancy is resolved below.

## Scope limitations

- New iterator progress guards detect an empty or unchanged next cursor. They do not deduplicate previously yielded items or detect arbitrary multi-page cursor cycles. A repeated page may be yielded once before termination, matching the existing member/thread iterator convention.
- Discovery sampled each domain, documentation, tests and critical REST/session/Gateway paths. This cycle was not an exhaustive audit of every source line or dependency advisory.
- Runtime latency, CPU, memory usage and production throughput have no baseline. A small header-parsing benchmark now has a controlled definition; it shows 0 allocations and no speed improvement. It is not a production workload profile.

## Resolved in cycle 2026-11-14_04-21

- P2: Extremely large REST retry/reset numbers overflowed to negative durations. `retry_after:1e100` became 1 ms, bypassing the documented 30-second retry cap; headers could also disable exhausted-bucket waiting. Local tests reproduced both unwanted second requests. Shared saturating conversion now preserves a positive maximum delay, rejects NaN/nonpositive inputs, and retains cancellation. Full race suite passes.

## Resolved in cycle 2026-11-14_04-51

- P2: `CreateWebhook` counted bytes against an 80-character limit, rejecting names already accepted by `CreateWebhookWithAvatar` and `webhookBody`. It now forwards to the shared creation implementation with an empty avatar. Local boundary tests failed for 80 accented letters and 80 emoji before the fix; all 14 constructor/boundary combinations pass afterward, including invalid-input rejection and preserved request options/auth/body.

## Resolved in cycle 2026-11-14_03-47

- P1: Shared mutable message/audit iterator query produced confirmed data races under concurrent use and skipped earlier pages on subsequent traversals. Fixed with a per-traversal query copy.
- P1: Message, audit-log and guild iteration could keep fetching a repeated full page. Fixed with cursor progress checks, verified against bounded local server fixtures.
