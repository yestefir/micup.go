# Latest cycle — 2026-11-14_11-51

**BUG FIX VERIFIED.** New READY/invalid-session boundaries invalidate old voice joins with the existing ErrVoiceGrant. A consumed old grant cannot be stored after reset, and earlier caller cleanup cannot remove a replacement join. Resumable reconnects preserve pending joins; terminal shutdown behavior remains covered.

Validated fix commit: `902adb4981dd093fff05c0e9e1c92853eb3116c1`.

Before the fix, two deterministic consumed-result probes restored old grants, four empty/buffered probes missed invalidation and two local pending joins expired at their deadlines. All now pass. Resume control succeeds before/after. These are correctness observations, not latency measurements; reset-loop overhead is unprofiled and previous allocation costs remain recorded.

Validation: 182 top-level race tests pass, 2 optional voice skips, 0 failures; vet/build/modified-file formatting pass. Full formatting reports 110 existing CRLF files. Targeted voice/reset/shutdown checks pass five repetitions.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md); acquire the overlap lock before changes. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_11-51.md).
