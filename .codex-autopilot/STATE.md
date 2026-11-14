# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Starting revision: `462dcf7bab5acd8a36b926fa2449a6d01a30224c`; initial working tree was clean.
- Validated code commit: `c034858b481b50cc4b2e814e81f93001a93d1a44`.
- Verdict: **BUG FIX VERIFIED**. Message/audit iterators now isolate traversal state, and message/audit/guild iterators stop on empty or unchanged cursors.
- Validation: 145 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting fails on 124 pre-existing CRLF files; see known issues.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Investigate webhook request consistency: `ExecuteWebhook` sends bot Authorization while the newer token-based webhook methods set `anonymous`; basic creation counts bytes while avatar-based creation counts runes. Verify intended contracts and add focused regression tests before changing behavior.
2. Investigate retry-duration parsing at extreme numeric inputs in `errors.go`; inspect whether duration conversion can bypass retry caps. Establish a deterministic test before proposing a fix.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. Separately resolve the Windows CRLF formatting check without broad unrelated source changes.

These are investigation candidates, not confirmed platform/security defects. No measured runtime performance bottleneck has been established.
