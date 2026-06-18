# croni — A friendly CLI for managing scheduled jobs on macOS

croni wraps macOS `launchd` behind a cron-like interface. Schedule recurring commands — especially `claude -p "..."` prompts — that survive reboots, without ever touching plist XML or `launchctl`.

## Installation

### Homebrew

```bash
brew tap dsaiztc/tap
brew install croni
```

### From source (remote)

```bash
go install github.com/dsaiztc/croni@latest
```

### From source (local)

```bash
git clone https://github.com/dsaiztc/croni.git
cd croni
make build
# Binary at bin/croni
```

## Quick Start

```bash
# Run a command every 6 hours
croni add backup --every 6h --command "rsync -a ~/docs /backup/"

# Run at specific cron schedule (weekdays at 9am)
croni add standup-prep --cron "0 9 * * MON-FRI" --command "claude -p 'Summarize yesterday'"

# One-shot reminder in 20 minutes
croni add reminder --at 20m --command "say 'Time is up'"

# List all jobs
croni list

# Check logs
croni logs backup

# Force-run a job now
croni run backup

# Disable/enable without deleting
croni disable backup
croni enable backup

# Remove a job
croni remove backup
```

## Commands

| Command | Description |
|---------|-------------|
| `croni add <name>` | Create a new scheduled job |
| `croni list` | List all jobs |
| `croni show <name>` | Show details of a job |
| `croni edit <name>` | Modify an existing job |
| `croni run <name>` | Force-run a job now (synchronous) |
| `croni logs <name>` | View job logs (`--stderr`, `--tail N`, `--follow`) |
| `croni enable <name>` | Enable a disabled job |
| `croni disable <name>` | Disable a job (unloads from launchd) |
| `croni remove <name>` | Remove a job (`--force`, `--with-logs`) |
| `croni export <name>` | Dump generated launchd plist to stdout |

## Schedule Types

### Cron (`--cron`)

Standard 5-field cron expressions with full support for ranges, lists, steps, named months (JAN-DEC), named weekdays (MON-SUN), and special shortcuts:

```bash
croni add job --cron "*/15 9-17 * * MON-FRI" --command "..."   # Every 15min during work hours
croni add job --cron "@daily" --command "..."                    # Midnight daily
croni add job --cron "@hourly" --command "..."                   # Top of every hour
```

Supported shortcuts: `@yearly`, `@annually`, `@monthly`, `@weekly`, `@daily`, `@midnight`, `@hourly`

### Interval (`--every`)

Simple recurring interval:

```bash
croni add job --every 5m --command "..."    # Every 5 minutes
croni add job --every 2h --command "..."    # Every 2 hours
croni add job --every 1d --command "..."    # Every day
```

### One-shot (`--at`)

Fire once, then auto-remove:

```bash
croni add job --at 20m --command "..."                   # In 20 minutes
croni add job --at "2026-05-13T15:00" --command "..."    # At specific time
```

> **Note:** If the machine is asleep when the scheduled time passes, the job will not execute retroactively. Expired one-shot jobs are cleaned up automatically on the next `croni list`.

## Flags

### `croni add` flags

| Flag | Description |
|------|-------------|
| `--cron` | 5-field cron expression (one of cron/every/at required) |
| `--every` | Interval shorthand: `5m`, `2h`, `1d` |
| `--at` | One-shot: ISO 8601 or relative (`20m`, `2h`) |
| `--command` | Command to run (required) |
| `--workdir` | Working directory (defaults to `$PWD`) |
| `--env KEY=VAL` | Environment variables (repeatable) |
| `--description` | Human-readable description |
| `--disabled` | Create without loading into launchd |
| `--run-on-load` | Run immediately when enabled |

### Global flags

| Flag | Description |
|------|-------------|
| `--json` | JSON output on all commands |

## JSON Output

Every command supports `--json` for machine-readable output:

```bash
croni list --json
croni show my-job --json
croni add my-job --every 1h --command "echo hi" --json
# {"status":"ok","job":{...}}
```

## How It Works

croni manages jobs in `~/.croni/jobs.json` and generates macOS launchd plists under `~/Library/LaunchAgents/`. Each job runs via `croni _exec`, which handles logging and one-shot cleanup.

```
~/.croni/
├── jobs.json              # Source of truth for all job metadata
├── logs/                  # stdout/stderr logs per job
│   ├── my-job.stdout.log
│   └── my-job.stderr.log
└── generated/             # Copy of generated plists
```

## Development

```bash
make build    # Build binary with version info
make test     # Run tests with race detector
make vet      # Static analysis
make clean    # Remove build artifacts
```

## Releasing

Releases are automated via GoReleaser on tag push:

```bash
git tag v0.x.y
git push origin v0.x.y
```

This triggers `.github/workflows/release.yml`, which:
- Builds binaries for darwin (amd64/arm64)
- Creates a GitHub release with a changelog
- Publishes a formula to [dsaiztc/homebrew-tap](https://github.com/dsaiztc/homebrew-tap)

**Required GitHub secret**: `HOMEBREW_TAP_TOKEN` — a PAT with write access to the tap repo.
