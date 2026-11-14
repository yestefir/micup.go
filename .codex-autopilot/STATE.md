# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_12-51` (Europe/Istanbul).
- Latest starting revision: `a842c1c`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `13f3bd5efab39e8c1a9f66b6f5b92820ed5f5245`.
- Verdict: **BUG FIX VERIFIED**. Pending voice joins preserve and validate the latest grant received before commit. Successful commit atomically retires the reservation, so later updates use rotation/voice-loss callbacks instead of being swallowed before caller cleanup. Refusal, shutdown and replacement guards remain covered.
- Validation: 186 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting still fails on 110 remaining pre-existing CRLF files; see known issues.
- Performance: six buffered/consumed grant supersession probes previously committed older grants, now zero. Two post-commit update probes previously delivered zero callbacks, now two total. Buffered refusal control passes before/after. Correctness counts only; no latency/throughput/allocation comparison this cycle. Prior reservation/construction/snapshot costs remain explicit in history/known issues.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Inspect remaining REST/session cancellation and cleanup boundaries using bounded local fixtures; queued Gateway write access now respects cancellation, while active writes keep the existing five-second socket timeout.
2. Probe member chunk replay and completion ordering across session resume with bounded fixtures; pending voice rotation/commit handoff, shutdown and replacement ownership are resolved.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No production runtime bottleneck has been established. Prior iterator/retry/webhook fixes remain covered by the full race suite.
