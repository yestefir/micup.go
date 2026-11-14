# Latest cycle — 2026-11-14_07-21

**BUG FIX VERIFIED.** Cached channel/role snapshots now isolate nested pointers, collections and metadata. Five mutation groups failed before the fix; concurrent snapshot edits emitted actual data races. Regression/full race checks pass afterward.

Validated fix commit: `f299b78343833a89db91e76f0c8dc917e5d6653e`.

Validation: 163 top-level tests pass with race detection, 2 optional voice tests skip, 0 failures; vet and build pass. Changed Go files pass formatting. Full formatting still reports 116 existing CRLF files. Ownership has a measured cost: populated channel snapshots now use 1568 B/49 allocations (previously 48 B/1); populated roles use 96 B/7 allocations (previously 0). Minimal snapshots remain allocation-free. No speed improvement is claimed.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md) before the next invocation; acquire the overlap lock before changing files. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_07-21.md).
