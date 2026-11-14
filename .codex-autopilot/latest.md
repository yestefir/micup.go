# Latest cycle — 2026-11-14_09-21

**BUG FIX VERIFIED.** Missing upload keys/IDs, invalid part URLs and embedded URL credentials are rejected before file/data transfer. All eight new metadata/URL regressions failed before the fix and pass afterward; valid signed queries and previous upload cases pass. URL-supplied Authorization was observed only in synthetic fixtures; no bot-token leak or production incident is claimed.

Validated fix commit: `8d3dd2b6ad2685cdd415cc9c90a3a6321c052bc7`.

Validation: 171 top-level tests pass with race detection, 2 optional voice tests skip, 0 failures; vet and build pass. Changed Go files pass formatting. Full formatting still reports 114 existing CRLF files. No speed improvement is claimed; prior snapshot ownership costs remain in history/known issues.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md) before the next invocation; acquire the overlap lock before changing files. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_09-21.md).
