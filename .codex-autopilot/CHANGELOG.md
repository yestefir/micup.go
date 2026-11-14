# Engineering audit changelog

## 2026-11-14 — Cycle 2026-11-14_11-21

- Added member-inventory context/target checks before page lookups and cancellation/availability checks before completion. Workers stop obsolete pagination after READY, guild sync, guild deletion/outage or cancellation, while retaining existing epoch identity, retry and event-reconciliation rules.
- Eight bounded regressions fail before the fix and pass afterward. Four lifecycle probes make 4 total page calls versus 12; pre-cancelled/mid-page context probes make 1 versus 4. Short outage/cancelled last pages no longer publish Complete. Custom in-flight lookups must honor context themselves.
- Validation: 178 top-level race tests pass, 2 optional voice skips, no failures; targeted inventory/shutdown regressions pass five repetitions. Vet/build/modified-file formatting/whitespace pass. Full formatting reports 110 existing CRLF files after formatting the modified inventory source.
- No speed/latency claim or allocation comparison; added per-page mutex check overhead is unprofiled and prior constructor/snapshot costs remain recorded.
- Fix commit: `aeafe8e2cf419bcace1bcce0ab835de66c2d00f5`. No push, deployment, dependencies, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_10-51

- Fixed cancelled shard IDENTIFY calls reserving future cooldown slots. Admission now remains serialized through the cooldown, supports context cancellation while queued and advances the next deadline only on success; already admitted cooldowns are retained.
- Reproduced 20 seconds of extra scheduled delay from four cancellations in both pre-cancelled and mid-wait probes; each now adds zero. Added busy-admission cancellation coverage; full shard race tests pass five repetitions, including normal pacing and fatal shutdown.
- Limiter-construction allocation increases from 48 B/1 to 160 B/2 per manager. Three observed ns/op samples: 29.52/30.04/28.29 before; 76.64/77.75/79.11 after. This deliberate reliability cost is not a speed improvement; schedule-state checks are not production latency measurements.
- Validation: 175 top-level race tests pass, 2 optional voice skips, no failures; vet/build/modified-file formatting/whitespace checks pass. Full formatting reports 111 pre-existing CRLF files after formatting two modified shard files.
- Fix commit: `08e80df6225cedd3d5978efab57d294fc58fc407`. No push, deployment, dependencies, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_10-21

- Replaced the Gateway writer mutex with a context-selectable per-connection gate. Queued commands/heartbeats can return on cancellation/deadline while the active writer continues; shared compressor/socket writes remain serialized and contexts are rechecked inside the gate.
- Four local stalled-transport cases failed before the fix at a 500 ms watchdog and pass afterward. Checks also retain the active write, following sends, absent cancelled frames and unchanged heartbeat state. Existing queued-send assertions are preserved with updated gate setup.
- Constructor benchmark (three runs/version): allocation increases from 160 B/3 to 272 B/4 per connection. Observed times 86.59/89.35/93.65 ns before and 128.4/128.2/131.7 ns after; no throughput or production latency claim. This is a deliberate correctness cost, not a speed improvement.
- Validation: 173 top-level race tests pass, 2 optional voice skips, no failures; targeted Gateway tests pass five repetitions. Vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 113 existing CRLF files.
- Fix commit: `98c815ef6f488e45e6ef549e89012f2f006cb317`. No push, deployment, dependencies, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_09-51

- Removed terminal REST backoffs after the fifth/final transport or retryable API failure; retry count and final rate-limit pacing remain intact. Documented exhaustion behavior.
- Regression reproduced cancellation masking the last global 429 error before the fix; final API error and pacing assertions pass afterward.
- Controlled exhaustion benchmarks, one iteration per mode/version: HTTP 503 15.5026391 -> 7.5032818 seconds (51.60% lower); transport error 15.5034012 -> 7.5023466 seconds (51.61% lower). Five attempts each, synthetic in-process transport, real backoff sleeps. No production performance extrapolation.
- Validation: 172 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting reports 113 existing CRLF files.
- Fix commit: `6f62fb8c7ff917c2c35e0c011f62b0531af68267`. No push, deployment, dependencies, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_09-21

- Added upload key/multipart ID checks and preflight validation of every multipart URL before file consumption or data PUTs. A shared URL validator retains HTTP(S)/host checks and rejects embedded credentials in either upload mode.
- Eight invalid metadata/URL cases previously read file bytes; five returned nil under permissive fixtures, and two credential-bearing URL cases sent an Authorization header. All now reject before file/data transfer. No bot-token leak or production incident is asserted.
- Added signed-query preservation coverage and retained numbered layouts, short-read handling, normal singlepart/multipart behavior and Authorization omission checks. The fixture's credential warning now accurately describes any Authorization header rather than calling URL-supplied auth a bot token.
- Validation: 171 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 114 existing CRLF files. No runtime speed claim.
- Fix commit: `8d3dd2b6ad2685cdd415cc9c90a3a6321c052bc7`. No push, deployment, dependency, public API or automation changes.

## 2026-11-14 — Cycle 2026-11-14_08-51

- Fixed multipart byte routing for numbered plans listed out of order. Parts now upload in numeric order, with exact expected count and consecutive 1-based numbering checked before file reads/data PUTs. Count-error messages are preserved; buffers and completion payloads are unchanged.
- Reproduced wrong part bodies in a reversed local plan. Six of eight malformed-layout fixtures previously consumed/uploaded data; four completed under the permissive fixture. All structural cases now fail before file consumption/data uploads. Added short-reader/no-completion regression coverage and documentation.
- Verified upstream numbering/part-size descriptions against the [official Fluxer upload guide](https://docs.fluxer.com/topics/uploads/). Sorting is defensive normalization; no claim that real instances return reversed plans or have suffered corruption.
- Validation: 169 top-level tests pass with race detection, 2 optional voice skips, no failures; vet/build/changed-file formatting/whitespace checks pass. Full formatting still reports 114 existing CRLF files. No runtime speed claim.
- Fix commit: `de3ea3e8bdf7b64687f8a91d3b1d4d134f65493f`. No push, deployment, dependency, public API or automation changes.

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
