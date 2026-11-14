# Latest cycle — 2026-11-14_04-21

**BUG FIX VERIFIED.** Fixed overflow in REST retry and rate-limit reset durations. Oversized delays previously became negative, bypassed the retry cap and allowed requests into an exhausted bucket. Regression tests fail against the baseline source and pass with saturating conversion.

Validated fix commit: `cb8ea97074fc5653614fff54390ab9d760c39735`.

Validation: 149 top-level tests pass with race detection, 2 optional voice tests skip, 0 failures; vet and build pass. Changed Go files pass formatting. The full formatting check still reports 123 existing CRLF files. Parsing benchmarks show 0 allocations in both versions and slightly slower observed medians (normal +2.79%, oversized +0.30%); sample ranges overlap. No speed improvement is claimed.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md) before the next invocation; acquire the overlap lock before changing files. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_04-21.md).
