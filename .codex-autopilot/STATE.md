# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_05-51` (Europe/Istanbul).
- Latest starting revision: `ac36740`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `6465cab84e7359d4351d3f6e4aab08720cf9a96b`.
- Verdict: **BUG FIX VERIFIED**. A saturated member-response buffer now fails and retires its request instead of silently losing data or its completion/error notification.
- Validation: 155 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting fails on 119 remaining pre-existing CRLF files; see known issues.
- Performance: this cycle measures deterministic overflow/error behavior, not runtime speed. Earlier header/snapshot benchmarks remain unchanged, including the documented 80-byte custom-snapshot ownership cost. No performance improvement claimed.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Inspect REST upload size/plan boundaries using synthetic streams and local responses; establish failures before changing allocation or streaming behavior.
2. Reproduce pending-member-request behavior at terminal session shutdown, distinguishing it from transient/resumable disconnects before changing cleanup semantics.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No production runtime bottleneck has been established. Prior iterator/retry/webhook fixes remain covered by the full race suite.
