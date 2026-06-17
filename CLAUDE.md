# croni

Go CLI for managing scheduled jobs on macOS. Wraps launchd behind a cron-like interface. Uses Cobra for commands, stores jobs in `~/.croni/jobs.json`, generates launchd plists.

## Build & Test

```bash
make build        # Build bin/croni with version info
make test         # Run tests with race detector
make vet          # Static analysis
```

After any change, run `make build && make test` to verify.

## Project Structure

- `main.go` — Entrypoint, calls `cmd.Execute()`
- `cmd/root.go` — Cobra root command, version/commit/date vars (set via ldflags), global `--json` flag
- `cmd/add.go` — Create a job (`--cron`, `--every`, or `--at` schedule)
- `cmd/edit.go` — Modify an existing job's fields
- `cmd/remove.go` — Delete a job (confirms in TTY, requires `--force` otherwise)
- `cmd/enable.go` / `cmd/disable.go` — Toggle launchd loading
- `cmd/list.go` — Tabular or JSON listing of all jobs
- `cmd/show.go` — Detail view of a single job
- `cmd/run.go` — Synchronous manual execution (stdout/stderr to terminal)
- `cmd/exec.go` — Hidden `_exec` command invoked by launchd (logs to files, auto-removes one-shots)
- `cmd/export.go` — Dump the generated plist XML to stdout
- `cmd/logs.go` — View stdout/stderr logs (`--follow`, `--tail`)
- `internal/types/job.go` — Job, Schedule, Store structs (JSON-serialized)
- `internal/store/store.go` — File-based JSON store with flock locking
- `internal/launchd/plist.go` — Plist generation, interval/at parsing, name validation
- `internal/launchd/ctl.go` — launchctl bootstrap/bootout/print wrappers
- `internal/cron/expand.go` — Cron expression to launchd CalendarInterval expansion

## Conventions

- **Schedule types**: exactly three — `cron` (5-field), `every` (interval shorthand like 5m/2h/1d), `at` (ISO 8601 or relative). Every add/edit must accept exactly one.
- **Cron features**: supports @daily/@hourly/etc. shortcuts, named months (JAN-DEC), named weekdays (MON-SUN), ranges, lists, steps.
- **Launchd label prefix**: all plist labels use `com.croni.<name>`. Do not change this prefix.
- **Storage**: `~/.croni/jobs.json` is the single source of truth. All mutations go through `store.withLock()` to prevent races.
- **Plist lifecycle**: add = generate + InstallPlist + Bootstrap. edit = Bootout + regenerate + InstallPlist + Bootstrap. remove = Bootout + RemovePlist + store.Remove. disable = Bootout + RemovePlist. enable = generate + InstallPlist + Bootstrap.
- **Exec path**: launchd runs `croni _exec --job <name>`, which looks up the job, runs it via bash, and appends stdout/stderr to `~/.croni/logs/`.
- **JSON output**: global `--json` flag on all commands. Mutating commands emit `{"status":"ok","job":{...}}`. Errors emit `{"status":"error","error":"..."}` to stderr.
- **Error handling**: return errors from RunE; root.go prints to stderr and exits 1.
- **Non-interactive safety**: `remove` without `--force` refuses when stdin is not a TTY.
- **JSON tags**: use snake_case (see types/job.go).
- **Cron OR semantics**: when both day-of-month and day-of-week are set, expand to two separate CalendarInterval sets (cron OR behavior vs launchd AND behavior).

## Release Workflow

1. Ensure `make vet` and `make test` pass
2. Commit all changes and push to main
3. Tag the release: `git tag vX.Y.Z`
4. Push the tag: `git push origin vX.Y.Z`
5. Wait for the GitHub Actions workflow to complete (builds binaries, updates Homebrew tap)
6. Edit the release notes: `gh release edit vX.Y.Z --notes "..."`

## Release Notes Format

Follow [Keep a Changelog](https://keepachangelog.com/) conventions. Start with a one-line summary, then use `###` sections as applicable:

- **Added** — new features
- **Changed** — changes to existing functionality
- **Fixed** — bug fixes
- **Removed** — features that were removed

Only include sections that apply to the release.
