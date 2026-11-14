# Latest cycle — 2026-11-14_12-21

**BUG FIX VERIFIED.** A voice grant consumed before terminal shutdown can no longer be stored afterward. Retained reservations carry closure metadata, so completion returns existing ErrGatewayClosed while same-guild exclusion lasts until caller cleanup. Fresh reservations and replacement/resume behavior remain covered.

Validated fix commit: `502a9c59e1585c4f800eabb95f9fed9709b184bf`.

Both empty/buffered result-queue probes fail before the fix and pass afterward. The Session.New + 512 reservations benchmark has median 196840 -> 220764 B/op (+12.154%); 1058 allocations unchanged. Observed time median also rises (123881 -> 134006 ns/op, three samples/version). This construction workload is not production latency, retained memory or a broad performance trend; prior costs remain recorded.

Validation: 183 top-level race tests pass, 2 optional voice skips, 0 failures; vet/build/modified-file formatting pass. Full formatting reports 110 existing CRLF files. Targeted voice/shutdown checks pass five repetitions.

Active recurring automation: `micup-go-engineering-autopilot`, every 30 minutes in this chat. Read the [runbook](RUNBOOK.md) and [state](STATE.md); acquire the overlap lock before changes. No pushes or deployments.

Full evidence and next priorities: [execution report](reports/2026-11-14_12-21.md).
