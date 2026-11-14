# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_11-21` (Europe/Istanbul).
- Latest starting revision: `2eaaef0`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `aeafe8e2cf419bcace1bcce0ab835de66c2d00f5`.
- Verdict: **BUG FIX VERIFIED**. Member inventory workers stop fetching further pages after lifecycle invalidation/cancellation, and recheck cancellation/availability before publishing completion. Eight bounded regression cases fail before the fix and pass afterward.
- Validation: 178 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting still fails on 110 remaining pre-existing CRLF files; see known issues.
- Performance: four bounded obsolete-lifecycle probes now make 4 total page lookups versus 12; two context probes make 1 versus 4. Two obsolete last-page cases no longer publish Complete. These are controlled callback counts/correctness observations, not production request counts or latency measurements. Per-page mutex check overhead is unprofiled; earlier constructor/snapshot costs remain explicit.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Inspect remaining REST/session cancellation and cleanup boundaries using bounded local fixtures; queued Gateway write access now respects cancellation, while active writes keep the existing five-second socket timeout.
2. Probe voice grant delivery across new READY/invalid-session boundaries before changing join ownership; inventory lifecycle invalidation and cancelled IDENTIFY reservations are resolved.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No production runtime bottleneck has been established. Prior iterator/retry/webhook fixes remain covered by the full race suite.
