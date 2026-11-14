# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_04-21` (Europe/Istanbul).
- Latest starting revision: `3aed5c1`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `cb8ea97074fc5653614fff54390ab9d760c39735`.
- Verdict: **BUG FIX VERIFIED**. Oversized REST retry/reset durations now saturate instead of overflowing to negative waits; the documented 30-second retry cap and context cancellation work for these inputs.
- Validation: 149 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting fails on 123 remaining pre-existing CRLF files; see known issues.
- Performance: five serial 200 ms benchmark samples per input before/after; 0 allocations in both versions. Median normal parsing 29.39 to 30.21 ns/op (2.79% slower observed), oversized parsing 36.32 to 36.43 ns/op (0.30% slower observed). Overlapping ranges; no speed improvement claimed.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Investigate webhook request consistency: `ExecuteWebhook` sends bot Authorization while the newer token-based webhook methods set `anonymous`; basic creation counts bytes while avatar-based creation counts runes. Verify intended contracts and add focused regression tests before changing behavior.
2. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests.
3. Resolve the Windows CRLF formatting check through a focused checkout-policy change without broad unrelated source edits.

The webhook inconsistencies are investigation candidates, not confirmed platform/security defects. No measured runtime performance bottleneck has been established. Prior iterator fixes remain covered by the full race suite; see the earlier report for their evidence.
