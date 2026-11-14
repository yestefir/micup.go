# Engineering audit changelog

## 2026-11-14 — Cycle 2026-11-14_03-47

- Fixed `EachMessage` and `EachAuditLogEntry` sharing cursor state across traversals. New tests failed before the fix, and race detection reported conflicting accesses at original `iterate.go:43`, `:45` and `:121`. Each traversal now starts from an immutable captured query.
- Fixed repeated-page request loops in `EachMessage`, `EachAuditLogEntry` and `EachGuild` by stopping when a cursor is empty or unchanged. Bounded pre-fix fixtures reached their third-request cutoff; fixed versions terminate after two requests. This is a correctness result, not a throughput or latency benchmark.
- Added reusable/concurrent traversal, stalled-cursor and normal guild pagination regression coverage. Updated message/guild guides to describe traversal and stop behavior.
- Validation: 145 top-level test passes under `go test -race -json ./...`, 2 optional voice skips, no failures; vet and build passed. Changed Go formatting and `git diff --check` passed. Full formatting remains affected by 124 pre-existing CRLF files.
- Established durable audit state, report, measurement history, original instructions and a cycle lock protocol. Activated the requested 30-minute app automation. No push or deployment.
