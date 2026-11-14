# Latest cycle — 2026-11-14_05-21

**BUG FIX VERIFIED.** Session presence now isolates nested custom status at construction, update and getter boundaries. All three caller-mutation regressions failed before the fix; concurrent returned-snapshot edits reproduced a DATA RACE. Targeted and full race checks pass with the fix.

Validated fix commit: `a0ee049f06baa44701a1afd4c64bce23e65f5869`.

Validation: 153 top-level tests pass with race detection, 2 optional voice tests skip, 0 failures; vet and build pass. Changed Go files pass formatting. Full formatting still reports 120 existing CRLF files. Retaining a custom-status snapshot costs 80 extra bytes and one allocation; median microbenchmark time rises from 32.18 to 70.26 ns/op. This is a measured ownership/correctness tradeoff, not a speed improvement.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md) before the next invocation; acquire the overlap lock before changing files. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_05-21.md).
