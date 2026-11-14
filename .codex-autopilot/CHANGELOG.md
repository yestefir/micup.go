# Engineering audit changelog

## 2026-11-14 — Cycle 2026-11-14_05-21

- Fixed session custom-status aliasing in initial options, `Presence()` snapshots and `UpdatePresence` input. A private nested copy helper isolates caller-owned and session-owned values while retaining presence validation, helpers and Gateway behavior.
- Three mutation regressions failed before the fix; concurrent edits to eight returned snapshots emitted an actual DATA RACE. Targeted and full race tests now pass. Added snapshot ownership guidance to the Gateway docs.
- Snapshot benchmarks (5 serial 200 ms samples/version) show plain 33.23 to 33.26 ns/op median, unchanged 32 B/1 allocation; custom 32.18 to 70.26 ns/op, 32 to 112 B, 1 to 2 allocations. The extra custom-status copy has an explicit correctness cost, not a speed claim.
- Validation: 153 top-level tests pass with race detection, 2 optional voice tests skip, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 120 unrelated CRLF files.
- Fix commit: `a0ee049f06baa44701a1afd4c64bce23e65f5869`. No push, deployment, production data/dependency or automation changes.

## 2026-11-14 — Cycle 2026-11-14_04-51

- Fixed `CreateWebhook` rejecting valid multibyte names by forwarding to the existing `CreateWebhookWithAvatar` implementation with an empty avatar. The two methods now share the established 1-to-80-code-point limit and optional avatar omission.
- Added 14 local constructor/boundary combinations: 80 ASCII letters, accented letters and emoji; overlong ASCII/Unicode names; empty name; missing channel. Invalid inputs make no request; accepted inputs retain name, route, bot auth and audit reason. The two basic Unicode boundary cases failed before the fix and pass afterward.
- Inspected the authorization candidate: an existing test explicitly requires bot Authorization on webhook execution. Preserved that tested behavior and deferred any change until new contract evidence exists.
- Validation: 150 top-level tests pass under race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting reports 121 remaining pre-existing CRLF files. No speed claim or runtime performance comparison.
- Fix commit: `8ac963e4149972b28b96228c39a41ea8e3ea2724`. No push, deployment, dependency change or automation edit.

## 2026-11-14 — Cycle 2026-11-14_04-21

- Fixed REST delay overflow in `errors.go`: positive oversized body/header/reset durations saturate at the maximum `time.Duration` instead of becoming negative. This restores the existing retry cap and exhausted-bucket waits. NaN and nonpositive inputs yield zero; valid fractional durations and the body's 1 ms floor are preserved.
- Added four regression tests covering numeric boundaries, body/header decoding, unwanted retries and bucket cancellation. All bug scenarios failed against baseline source before passing with the fix.
- Added `BenchmarkHeaderSeconds`, five 200 ms samples per input/version on the same machine, run serially. Median parsing time was 29.39 to 30.21 ns/op for `0.25` and 36.32 to 36.43 ns/op for `1e100`; both versions allocate 0 B/op and 0 allocs/op. Slightly slower observed medians with overlapping ranges; no speed improvement claimed.
- Validation: 149 top-level tests pass under race detection, 2 optional voice tests skip, 0 failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 123 unrelated CRLF files.
- Fix commit: `cb8ea97074fc5653614fff54390ab9d760c39735`. No push or deployment; recurring automation unchanged.

## 2026-11-14 — Cycle 2026-11-14_03-47

- Fixed `EachMessage` and `EachAuditLogEntry` sharing cursor state across traversals. New tests failed before the fix, and race detection reported conflicting accesses at original `iterate.go:43`, `:45` and `:121`. Each traversal now starts from an immutable captured query.
- Fixed repeated-page request loops in `EachMessage`, `EachAuditLogEntry` and `EachGuild` by stopping when a cursor is empty or unchanged. Bounded pre-fix fixtures reached their third-request cutoff; fixed versions terminate after two requests. This is a correctness result, not a throughput or latency benchmark.
- Added reusable/concurrent traversal, stalled-cursor and normal guild pagination regression coverage. Updated message/guild guides to describe traversal and stop behavior.
- Validation: 145 top-level test passes under `go test -race -json ./...`, 2 optional voice skips, no failures; vet and build passed. Changed Go formatting and `git diff --check` passed. Full formatting remains affected by 124 pre-existing CRLF files.
- Established durable audit state, report, measurement history, original instructions and a cycle lock protocol. Activated the requested 30-minute app automation. No push or deployment.
