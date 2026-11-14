# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_07-21` (Europe/Istanbul).
- Latest starting revision: `60c8aec`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `f299b78343833a89db91e76f0c8dc917e5d6653e`.
- Verdict: **BUG FIX VERIFIED**. Cached channel/role getters now return independent nested snapshots; five mutation groups and concurrent metadata/color races were reproduced before the fix and pass afterward.
- Validation: 163 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting still fails on 116 remaining pre-existing CRLF files; see known issues.
- Performance: five serial 200 ms samples/version for minimal/populated snapshots. Channel median 93.98 to 105 ns/op minimal (0 allocations), 131.6 to 1812 ns/op populated (48 to 1568 B; 1 to 49 allocations). Role median 45.60 to 53.36 ns/op minimal (0 allocations), 45.23 to 231.7 ns/op populated (0 to 96 B; 0 to 7 allocations). This ownership fix has measured cost; no speed improvement claimed.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Inspect REST upload size/plan boundaries using synthetic streams and local responses; establish failures before changing allocation or streaming behavior.
2. Inspect ownership of caller-supplied session/shard configuration slices; establish a supported mutation/race scenario before changing copying behavior.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No production runtime bottleneck has been established. Prior iterator/retry/webhook fixes remain covered by the full race suite.
