# Latest cycle — 2026-11-14_12-51

**BUG FIX VERIFIED.** Pending voice joins use the latest grant received before commit and validate that result. Successful commit retires the reservation with the cache write, so subsequent grants use rotation/voice-loss callbacks even before the original caller's deferred cleanup.

Validated fix commit: `13f3bd5efab39e8c1a9f66b6f5b92820ed5f5245`.

Six buffered/consumed supersession cases and two post-commit callback cases fail before the fix and pass afterward. Buffered refusal control passes before/after; shutdown/replacement/resume guards remain covered. These are correctness observations, not latency measurements. Prior reservation/construction/snapshot costs remain recorded; new mutex-scope overhead is unprofiled.

Validation: 186 top-level race tests pass, 2 optional voice skips, 0 failures; vet/build/modified-file formatting pass. Full formatting reports 110 existing CRLF files. Targeted voice checks pass five repetitions.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md); acquire the overlap lock before changes. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_12-51.md).
