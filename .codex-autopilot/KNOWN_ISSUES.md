# Known issues and investigation candidates

## Confirmed unresolved validation gaps

| Priority | Issue | Impact and evidence | Next action |
|---|---|---|---|
| P2 | Optional voice integration checks skipped locally | `TestPCMTrackEncodesOpus` skips because FFmpeg is absent; `TestLocalLiveKitReceivesOpusAudio` skips because `LIVEKIT_TEST_BINARY` is unset. The remaining tests pass, but real audio encoding/delivery is not verified here. | Use local test dependencies when available; rerun those tests without production services. |
| P3 | Full formatting check fails on 110 remaining existing CRLF files | `core.autocrlf=true`; initial cycle verified CRLF-only formatting differences. Count fell by one after formatting the modified inventory file; modified Go files pass formatting and unrelated files are untouched. | Consider a focused LF attributes/checkout policy change, separately from behavior fixes. Do not claim `make check` passed. |

## Unverified candidates

- `ExecuteWebhook` sends bot Authorization while token-based lookup/edit methods omit it. The existing `TestWebhookEndpoints` explicitly requires that execution behavior. No actual authorization failure has been reproduced; retain the tested contract and reopen only with new upstream evidence. The creation-name discrepancy is resolved below.

## Scope limitations

- Member inventory workers recheck context and target identity/availability at page boundaries and completion. They do not forcibly interrupt an already-started custom MemberLookup call; that implementation must honor its context. A lifecycle/context change just after admission may overlap the in-flight lookup. Added per-page mutex check overhead has not been profiled; lookup-count fixtures do not establish production speed.

- Shard IDENTIFY waiting now uses one admission channel per manager. The limiter constructor costs 160 B/2 allocations versus 48 B/1 allocation before; observed constructor times increased. Cancelled waits do not extend cooldowns, but a call already admitted before cancellation still consumes its interval. No strict shard ordering or end-to-end wire spacing guarantee is added; dialing/scheduling follow admission.

- Gateway queued write-access waits now end with context cancellation/deadlines. Already-started writes are not actively interrupted by caller cancellation and retain the five-second socket write timeout. Serialization before queue admission is also synchronous. The local stalled-writer regression checks a 500 ms response bound, not production latency. A per-connection channel adds 112 B and one constructor allocation (272 B/4 versus 160 B/3); constructor benchmark time also increased. Production effect is unmeasured.
- New iterator progress guards detect an empty or unchanged next cursor. They do not deduplicate previously yielded items or detect arbitrary multi-page cursor cycles. A repeated page may be yielded once before termination, matching the existing member/thread iterator convention.
- Discovery sampled each domain, documentation, tests and critical REST/session/Gateway paths. This cycle was not an exhaustive audit of every source line or dependency advisory.
- Production-service latency, CPU, process memory usage and production throughput have no baseline. Controlled retry-exhaustion elapsed benchmarks use real sleeps and an in-process transport, one iteration per mode/version; they do not establish production latency distributions. Header/parsing/presence/cache microbenchmarks have controlled definitions, not production profiles. Independent custom presence snapshots cost an extra 80 B and one allocation. Fully populated cache-channel snapshots cost 1568 B/49 allocations versus the prior unsafe 48 B/1 allocation; populated role snapshots cost 96 B/7 allocations versus 0. Minimal channel/role snapshots still allocate zero. These ownership costs are measured and explicit; production impact is not established.

## Resolved in cycle 2026-11-14_07-21

- P1: Channel getters cloned only permission overwrites; role getters shallow-copied all fields. Newer pointer/map/slice metadata therefore remained shared with cached state and other callers. Four channel mutation groups plus role metadata mutation failed before the fix, and concurrent edits of separate snapshots emitted DATA RACE traces for thread metadata and role colors. Explicit private copies now isolate all current mutable references, including nested users/members/forum tags/role colors. Nil fields are preserved; mutation and full race checks pass. Added a fixture-completeness guard to highlight future reference fields requiring copy review.

## Resolved in cycle 2026-11-14_07-51

- P1: `Session.New` retained the caller's `IgnoredEvents` backing array after validation. Editing/reusing it after construction changed session filtering and could make previously valid Gateway options fail validation. Concurrent caller edits/session reads reproduced an actual DATA RACE. Constructor now uses `slices.Clone`; shard-created sessions inherit the fix, while shard ID inputs were already copied. Both regressions fail before the fix and pass afterward.

## Resolved in cycle 2026-11-14_08-21

- P2: `Conn.Send` checked context only before serialization/queueing. A send cancelled while waiting for write access still returned nil and reached the local Gateway as an opcode 4 command or heartbeat. Rechecking after write-lock acquisition now returns the context error before compression, heartbeat timing changes or writing. Both wire-level regression cases failed before the fix and pass afterward; a following active command still arrives normally.

## Resolved in cycle 2026-11-14_08-51

- P2: Multipart file segments were assigned by response-array order instead of `part_number`. A reversed local plan sent `ef` to part 1 and `abcd` to part 2 for the six-byte fixture. Count-mismatched layouts read/uploaded partial data before failing, and duplicate/nonpositive/gapped numbering could complete under a permissive local server. Client now sorts by number and validates count/consecutive 1-based coverage before reads/data PUTs. Two valid-order cases, eight invalid layouts and a short-reader case pass. No production corruption incident or upstream response-order guarantee is claimed.

## Resolved in cycle 2026-11-14_09-21

- P2: Missing upload keys/IDs and invalid later part URLs were not fully checked before consuming/uploading data; missing keys could return an unusable successful attachment under local fixtures. Embedded URL credentials also generated an Authorization header on upload requests, contrary to the upload transport contract; this was URL-supplied Basic auth, not a demonstrated bot-token leak. Plans now require nonempty keys/multipart IDs, preflight all multipart URLs and reject userinfo. Eight local metadata/URL cases fail before the fix and pass afterward without file/data side effects. Valid signed query strings are preserved.

## Resolved in cycle 2026-11-14_05-21

- P1: Session presence shallow copies leaked the nested `CustomStatus` pointer through constructor input, getter snapshots and `UpdatePresence` input. Caller edits silently changed stored state, bypassing update validation, and eight callers editing their returned snapshots reproduced a DATA RACE. A shared private copy helper now isolates all three boundaries; regression and full race tests pass.

## Resolved in cycle 2026-11-14_05-51

- P1: `deliverChunk` silently dropped responses when a request's 16-slot queue filled. Dropping a middle chunk could allow incomplete success after a later completion, while dropping completion/error could leave a request awaiting its context. Controlled middle/final overflow tests failed before the fix. Overflow now retires the request and reserves a queue slot for an explicit incomplete-list error under the existing mutex. Later responses are ignored; normal ordered responses and other requests remain intact.

## Resolved in cycle 2026-11-14_06-21

- P2: When `Session.Run` stopped, pending member requests with independent caller contexts remained registered and waited for their own deadlines. A local Gateway test reproduced two requests returning deadline exceeded after the run had stopped. Final run cleanup now removes registrations, drains buffered responses and notifies each request with the existing `gateway.ErrClosed`/public `micup.ErrGatewayClosed`. Empty/full-queue tests and the public request fixture pass. Transient reconnects do not execute this final cleanup.

## Resolved in cycle 2026-11-14_06-51

- P2: Pending voice joins with independent caller contexts also waited until their own deadlines after `Run` stopped. A local Gateway fixture held a confirmed join response and reproduced deadline exceeded after terminal shutdown. Final cleanup now signals `gateway.ErrClosed`/public `micup.ErrGatewayClosed`, replaces buffered responses safely under the notification mutex, and clears refusal ordering. Guild reservations remain until callers execute their existing deferred `endJoin`, preventing premature guild reuse. Normal joins, refusals, resumption and member cleanup pass regression/race checks.

## Resolved in cycle 2026-11-14_04-21

- P2: Extremely large REST retry/reset numbers overflowed to negative durations. `retry_after:1e100` became 1 ms, bypassing the documented 30-second retry cap; headers could also disable exhausted-bucket waiting. Local tests reproduced both unwanted second requests. Shared saturating conversion now preserves a positive maximum delay, rejects NaN/nonpositive inputs, and retains cancellation. Full race suite passes.

## Resolved in cycle 2026-11-14_04-51

- P2: `CreateWebhook` counted bytes against an 80-character limit, rejecting names already accepted by `CreateWebhookWithAvatar` and `webhookBody`. It now forwards to the shared creation implementation with an empty avatar. Local boundary tests failed for 80 accented letters and 80 emoji before the fix; all 14 constructor/boundary combinations pass afterward, including invalid-input rejection and preserved request options/auth/body.

## Resolved in cycle 2026-11-14_03-47

- P1: Shared mutable message/audit iterator query produced confirmed data races under concurrent use and skipped earlier pages on subsequent traversals. Fixed with a per-traversal query copy.
- P1: Message, audit-log and guild iteration could keep fetching a repeated full page. Fixed with cursor progress checks, verified against bounded local server fixtures.

## Resolved in cycle 2026-11-14_09-51

- P2: REST retries slept after their fifth/final failure even though no further attempt could occur. For retryable server/transport failures this added an unnecessary eight-second backoff, and cancellation during the final wait could hide the final API error. Both paths now return immediately after the last failure. A five-response 429 regression failed before the fix with context canceled and now returns the final global API error while preserving limiter pacing. Controlled five-attempt server/transport benchmarks fell from about 15.50 to 7.50 seconds, one iteration per mode/version. No production latency distribution is claimed.
## Resolved in cycle 2026-11-14_10-21

- P2: Gateway sends used an unconditional mutex wait after serialization/rate admission, so context cancellation/deadlines could not return until an earlier writer finished. A real local WebSocket with a stalled client transport reproduced all four command/heartbeat cancellation/deadline cases still waiting 500 ms after their contexts ended. Replaced only the writer mutex with a per-connection context-selectable gate, retaining the inside-gate context recheck and shared compressor/socket serialization. All four cases now return while the active write remains stalled, transmit no cancelled frame, and preserve that writer and following sends. Constructor cost increases by 112 B/one allocation per connection; active writes retain their existing timeout.
## Resolved in cycle 2026-11-14_10-51

- P2: The shard IDENTIFY limiter reserved its next slot before waiting, and never reclaimed a slot when the context ended. Four already-cancelled calls and four calls cancelled while waiting each added 20 seconds to a seeded cooldown despite returning context.Canceled. This retained schedule can also affect a later manager Run. The limiter now serializes cancellable admission through the current cooldown and advances next only for successful admission, preserving earlier cooldowns. Both schedule regressions fail before the fix and pass afterward; busy-admission cancellation and existing multi-shard pacing/fatal shutdown tests pass. A per-manager channel adds 112 B/one limiter-construction allocation; no production restart latency is claimed.
## Resolved in cycle 2026-11-14_11-21

- P2: Member enumeration checked its target identity only after all pages, so READY/guild sync/deletion/outage still allowed later page lookups. Outages retained the target pointer and could publish the obsolete enumeration as Complete. A custom lookup finishing successfully as its context ended could likewise trigger further pages and completion. Four bounded lifecycle probes made three calls each before the fix and now make one; pre-cancelled/mid-page cancellation made two calls each and now make zero/one. Short terminal-page outage/cancellation probes also published Complete before the fix. Page-boundary context/lifecycle checks plus completion checks now stop obsolete work and keep interrupted inventories incomplete, preserving existing retry/reconciliation behavior. Eight cases fail before the fix and pass afterward. No production request count, unbounded-loop or forced in-flight cancellation claim.