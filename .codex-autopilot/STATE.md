# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_09-51` (Europe/Istanbul).
- Latest starting revision: `8560550`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `6f62fb8c7ff917c2c35e0c011f62b0531af68267`.
- Verdict: **VERIFIED IMPROVEMENT**. Exhausted REST retries return the final failure without a terminal backoff. Final rate-limit responses still update pacing; final API errors are no longer masked by cancellation during that redundant wait.
- Validation: 172 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting still fails on 113 remaining pre-existing CRLF files; see known issues.
- Performance: controlled five-attempt exhaustion benchmarks: HTTP 503 elapsed 15.5026391 -> 7.5032818 seconds (51.60% lower); transport failure 15.5034012 -> 7.5023466 seconds (51.61% lower). One iteration per mode/version, real backoff sleeps, in-process transport; no production latency/throughput inference. Earlier snapshot ownership costs remain explicit.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Measure remaining Gateway cancellation responsiveness during write-lock waits/in-progress writes before considering synchronization redesign; admission/recheck fixes do not interrupt those waits.
2. Inspect remaining REST/session cancellation and cleanup boundaries using bounded local fixtures; the measured terminal retry-backoff defect is resolved.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No production runtime bottleneck has been established. Prior iterator/retry/webhook fixes remain covered by the full race suite.
