# Engineering audit changelog

## 2026-11-14 — Cycle 2026-11-14_08-21

- Fixed cancelled queued Gateway sends being transmitted after write access became available. `Conn.Send` now rechecks context inside the existing write lock before compression/timing state/socket writes.
- Added controlled local WebSocket tests for a normal command and heartbeat. Baseline calls returned nil and emitted cancelled frames; fixed calls return `context.Canceled`, do not change heartbeat timing state, and preserve delivery of the next active command. Existing normal and compressed Gateway tests pass.
- Validation: 166 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 115 existing CRLF files. No speed claim; lock-wait/in-progress-write cancellation responsiveness remains a stated limitation.
- Fix commit: `6c94709a1966787375c5b36bff939dda621423df`. No push, deployment, dependency, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_07-51

- Fixed caller-owned ignored-event slice aliasing in `Session.New` with a construction-time clone. Later caller edits no longer alter stored filtering or invalidate future Gateway options. Shard manager construction already clones shard IDs and now inherits per-session ignored-event isolation.
- Added post-construction mutation/validation and concurrent caller-edit regression tests. Baseline stored options became invalid and emitted actual race traces; targeted/full race tests pass after the fix. Gateway docs describe ownership after construction.
- Validation: 165 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/root docs tests/changed-file formatting/whitespace checks pass. Full formatting still reports 116 existing CRLF files. No runtime speed claim or new timing trend.
- Fix commit: `c6e7e6569cb96588d2bb770dc1cbb5594f9f5720`. No push, deployment, dependency, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_07-21

- Fixed cached channel/role snapshot aliasing. Getters now copy nested pointers, slices and maps, including forum/thread metadata, recipients, thread members and role colors. Helpers update local getter copies by pointer to avoid extra aggregate copies; cache writes remain event-driven.
- Four channel mutation groups and role mutation checks failed before the fix. Concurrent edits of separate snapshots reproduced races in thread metadata and role colors. Added rich-fixture reference coverage, absent-field preservation and concurrent-edit regression coverage, plus ownership documentation.
- Benchmarks (5 serial 200 ms samples/version): minimal channel 93.98 to 105 ns/op; minimal role 45.60 to 53.36 ns/op, both 0 allocations. Populated channel 131.6 to 1812 ns/op, 48 to 1568 B/op, 1 to 49 allocations; populated role 45.23 to 231.7 ns/op, 0 to 96 B/op, 0 to 7 allocations. Accepted explicit ownership cost; no performance improvement claimed.
- Validation: 163 top-level race-test passes, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Whole formatting still reports 116 existing CRLF files.
- Fix commit: `f299b78343833a89db91e76f0c8dc917e5d6653e`. No push, deployment, dependency, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_06-51

- Fixed pending voice joins outliving terminal `Session.Run`. Final cleanup notifies each waiter with the existing closed-connection error and clears refusal ordering; reservations remain until each caller completes its existing deferred cleanup, avoiding premature guild reuse.
- Serialized voice grant/refusal queue writes with shutdown using the existing mutex. Empty/full queue notifications are bounded and channels are not closed. Transient reconnect handling and established voice grants are outside this change.
- Added a held-voice-response mode/counter to the local test fixture, a public pending-join regression and controlled empty/full-queue/ownership tests. Both regressions failed before the fix and pass afterward. Gateway docs describe the closed error.
- Validation: 159 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 117 existing CRLF files. No speed claim.
- Fix commit: `cf647d97d38bceda68e15eb74eb78cb502a437e6`. No push, deployment, dependency or automation changes.

## 2026-11-14 — Cycle 2026-11-14_06-21

- Fixed pending member requests outliving terminal session execution. `Session.Run` final cleanup now retires all registered member requests and queues the existing closed-connection error. Transient reconnects remain inside the run loop and retain their prior behavior.
- Cleanup drains bounded response queues before sending the error, avoiding a blocked terminal signal or incomplete success after eviction of buffered data. It uses the same mutex as delivery, does not close channels and adds no goroutines.
- Added a public local-Gateway regression with two independent caller contexts, plus empty/full queue shutdown coverage. Both tests failed before the fix; targeted and full race checks pass afterward. Gateway docs describe shutdown errors and partial results.
- Validation: 157 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Whole formatting still reports 119 existing CRLF files. No runtime speed claim.
- Fix commit: `f0549cc8be6ea17540c6e11989e18e3c81c828ce`. No push, deployment, dependency or automation edit.

## 2026-11-14 — Cycle 2026-11-14_05-51

- Fixed silent member-chunk loss on a full request response queue. Overflow now atomically retires the request and queues an explicit incomplete-list error in reserved capacity, preventing missing completion or false complete results after lost data.
- Added controlled middle/final overflow regressions, late-response rejection, independent-request checks and normal in-order delivery at the buffer boundary. Both overflow cases failed before the fix; targeted and full race checks pass afterward. Gateway docs now describe partial results with an overflow error.
- Validation: 155 top-level tests pass with race detection, 2 optional voice skips, 0 failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 119 pre-existing CRLF files. No speed improvement or production occurrence rate claimed.
- Fix commit: `6465cab84e7359d4351d3f6e4aab08720cf9a96b`. No push, deployment, dependency or automation changes.

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
