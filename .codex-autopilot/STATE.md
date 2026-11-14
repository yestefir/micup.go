# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_10-21` (Europe/Istanbul).
- Latest starting revision: `80253b2`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `98c815ef6f488e45e6ef549e89012f2f006cb317`.
- Verdict: **BUG FIX VERIFIED**. Gateway sends waiting for write access now return when their own contexts end, without waiting for or interrupting the active writer. Command/heartbeat cancellation and deadline cases pass with a locally stalled socket writer.
- Validation: 173 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting still fails on 113 remaining pre-existing CRLF files; see known issues.
- Performance: all four queued-context scenarios formerly remained blocked at the 500 ms watchdog; now all return while the active writer stays blocked. This is a bounded correctness check, not a latency distribution. The per-connection write gate increases constructor allocation from 160 B/3 allocations to 272 B/4 allocations; measured constructor time also increases (raw samples in report). No throughput or production speed claim. Prior retry-exhaustion timing and snapshot ownership costs remain in history.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Inspect remaining REST/session cancellation and cleanup boundaries using bounded local fixtures; queued Gateway write access now respects cancellation, while active writes keep the existing five-second socket timeout.
2. Review shard/session lifecycle handoffs and identify pacing for concrete shutdown or ownership defects before changing synchronization.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No production runtime bottleneck has been established. Prior iterator/retry/webhook fixes remain covered by the full race suite.
