# Latest cycle — 2026-11-14_10-21

**BUG FIX VERIFIED.** Gateway sends waiting for write access now return when their contexts end, while leaving the active writer running. Four command/heartbeat cancellation/deadline cases failed before the fix and pass afterward against a stalled local WebSocket transport. The cancelled frames are absent; active and following frames still arrive.

Validated fix commit: `98c815ef6f488e45e6ef549e89012f2f006cb317`.

The channel gate costs 112 B and one additional allocation per connection (272 B/4 versus 160 B/3); constructor benchmark time also increased. No production latency/throughput improvement is claimed. Active writes retain the existing five-second socket timeout.

Validation: 173 top-level tests pass with race detection, 2 optional voice tests skip, 0 failures; vet/build/changed-file formatting pass. Full formatting reports 113 existing CRLF files. Targeted Gateway cases pass five repetitions.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md); acquire the overlap lock before changes. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_10-21.md).
