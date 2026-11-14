# Latest cycle — 2026-11-14_07-51

**BUG FIX VERIFIED.** Session construction now owns its ignored-event slice. Mutating/reusing the caller's slice no longer changes filtering or invalidates stored Gateway configuration. Mutation and actual race regressions failed before the fix and pass afterward.

Validated fix commit: `c6e7e6569cb96588d2bb770dc1cbb5594f9f5720`.

Validation: 165 top-level tests pass with race detection, 2 optional voice tests skip, 0 failures; vet and build pass. Changed Go files pass formatting. Full formatting still reports 116 existing CRLF files. No speed improvement is claimed; prior snapshot ownership costs remain in history/known issues.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md) before the next invocation; acquire the overlap lock before changing files. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_07-51.md).
