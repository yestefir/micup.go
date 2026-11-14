# Current engineering state

- Repository: micup.go; scope is this checkout only.
- First cycle: `2026-11-14_03-47` (Europe/Istanbul).
- Latest cycle: `2026-11-14_04-51` (Europe/Istanbul).
- Latest starting revision: `b94310d`; initial working tree was clean and no external commits followed the preceding cycle.
- Latest validated code commit: `8ac963e4149972b28b96228c39a41ea8e3ea2724`.
- Verdict: **BUG FIX VERIFIED**. Basic webhook creation now shares the avatar variant's Unicode name validation; valid 80-code-point international names are accepted consistently.
- Validation: 150 top-level tests passed under the race detector, 2 optional voice tests skipped, 0 failures; vet and build passed. Changed Go files pass formatting. Whole-checkout formatting fails on 121 remaining pre-existing CRLF files; see known issues.
- Performance: this cycle addresses a validation bug and makes no speed claim. Earlier parsing benchmark samples and their slightly slower observed medians remain in history; no incomparable new timing was added.
- No production requests, load tests, pushes, deployments, dependency changes or public API additions/removals were made.
- Recurring automation: `micup-go-engineering-autopilot`, active every 30 minutes in this chat. Local execution requires the computer and app to remain running.
- Read `RUNBOOK.md`, `INSTRUCTIONS.md`, `latest.md` and `KNOWN_ISSUES.md` before the next cycle. Atomically claim `cycle.lock` before any modifications and release only your own lock.

## Architecture discovered

Public packages at the root and `voice/` are facades, enforced by AST tests. Implementation packages follow tested dependency layers: model/internal/gateway, events/rest/discovery, state, session, shard, bot. The voice package independently integrates LiveKit, WebRTC and FFmpeg. REST uses a shared request/retry/rate-limit layer and redirect-blocking HTTP client; iterator functions page lazily. Gateway owns a single WebSocket protocol loop, session owns reconnection/resumption and member enumeration, state uses a mutex-protected cache, and shard coordinates sessions and IDENTIFY pacing. No application database or HTTP server is part of this module. CI runs vet, build and race tests on Windows, Linux and macOS; Linux installs FFmpeg.

## Next priorities

1. Inspect Gateway/session cancellation and pending-request cleanup on disconnect; recent cycles focused on REST, so this offers a different reliability target.
2. Inspect REST upload size/plan boundaries using synthetic streams and local responses; establish failures before changing allocation or streaming behavior.
3. Close optional voice validation gaps with local FFmpeg/LiveKit when available; keep production credentials out of all tests. The low-priority Windows line-ending policy remains recorded separately.

The webhook name inconsistency is resolved. `TestWebhookEndpoints` explicitly requires bot Authorization for execution, so that behavior was retained; revisit only with new contract evidence or a failing supported scenario. No measured runtime performance bottleneck has been established. Prior iterator/retry fixes remain covered by the full race suite.
