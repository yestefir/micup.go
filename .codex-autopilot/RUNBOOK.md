# Engineering cycle operation

Read `INSTRUCTIONS.md` for the complete user request. The recurring automation runs every 30 minutes in this chat against `Y:\micup.go`; each invocation must finish within 30 minutes. The app and computer must remain running for local scheduled work. Scheduling is managed by the app, never by a repository script or an infinite loop.

## Scope and state

Only this repository is authorized. Read applicable repository instructions, `git status --short`, `STATE.md`, `latest.md`, `KNOWN_ISSUES.md` and recent commits before changing files. The first cycle mapped the architecture; later cycles should inspect changes and follow recorded priorities rather than repeat the whole discovery. Never overwrite unrelated edits. Do not push or deploy. Commit only completed, isolated changes after validation, documenting pre-existing failures separately.

## Exclusive cycle lock

Before any repository modification, atomically create `.codex-autopilot/cycle.lock` using exclusive creation. The file is ignored by Git. Its existence means another cycle may own the checkout. Do not truncate it or delete it based solely on its age. On collision, report the conflict in the chat and terminate without modifying files, including audit files.

For PowerShell, use `[IO.File]::Open` with `FileMode.CreateNew`, `FileAccess.Write` and `FileShare.None`. Write a unique execution ID and start timestamp, then close the handle. Keep the lock file until the entire cycle, report and optional commit are complete. Wrap work in `try/finally`; remove the lock only if its recorded execution ID matches your own. On an unexpected termination, a later invocation must inspect the owning run and establish that it has ended before recovering a stale lock. A 30-minute age alone does not prove that it is stale. The lock coordinates cooperating autopilot runs; it does not block manual edits or unrelated agents, so inspect Git again before staging.

## Validation and measurement

The project uses Go 1.26.8. `make check` requires formatting, vet, build and race tests. On Windows, equivalent individual commands are acceptable:

```powershell
gofmt -l .
go vet ./...
go build ./...
go test -race ./...
git diff --check
```

Treat nonempty `gofmt -l` output as a failed formatting check. The first checkout had 124 pre-existing CRLF-only formatting differences caused by `core.autocrlf=true`; do not silently suppress this or reformat unrelated files. Check changed Go files independently. FFmpeg and a local LiveKit binary are optional dependencies for tests; report skips explicitly. Do not use live Micup credentials or production services for validation.

For every regression, record the pre-fix failure and post-fix result when feasible. Preserve concise evidence rather than full logs. Temporary logs belong in ignored `scratch/` and should be removed before finishing. Never claim performance changes based on test or build timing with different cache states. Distinguish finite server-cutoff tests from unbounded production behavior. Use consistent workload identifiers in `PERFORMANCE_HISTORY.csv`; blank baseline/current values mean unavailable and blank improvement means not comparable.

## Reporting and time budget

Use Europe/Istanbul timestamps for report names (`YYYY-MM-DD_HH-mm.md`). Reserve at least five minutes of the 30-minute maximum for validation, reporting and cleanup. Update all required audit files using the A–H report format in `INSTRUCTIONS.md`. If validation cannot finish, restore only your own incomplete changes and preserve investigation findings. Finish without unnecessary code edits when no safe worthwhile fix is established.

Automation ID: `micup-go-engineering-autopilot` (active thread heartbeat, every 30 minutes). Do not recreate or reschedule it during routine runs.
