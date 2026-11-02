# taskgo

A local-first task and project tracker for developers, built for the terminal.

`taskgo` is designed around one idea: capturing work should never interrupt work. It keeps everything in a single SQLite database on your machine, infers the Project from the repository you are standing in, and reserves tidying up for a separate, deliberate act.

> **Status: design phase.** The domain model, ADRs, spec, and ticket breakdown are complete. The Go layout exists, but no command is implemented yet. The command surface below is the agreed v1 scope, not current behavior.

## Why another task tracker

- **Capture is one command.** No Project decision, no prompt, no network.
- **Repository-aware.** Inside a registered repository, new Tasks land in the matching Project automatically.
- **Global by default where it counts.** `next` considers everything you have open, not just the directory you are in.
- **Scriptable.** Every command supports `--json`, data goes to stdout, diagnostics to stderr, and exit codes are stable.
- **Yours.** One file, no account, no sync service, no background daemon.

## Install

Not yet released. To build from source:

```sh
make build        # produces bin/taskgo
./bin/taskgo      # currently a no-op entrypoint
```

Requires Go 1.27 or newer.

## Planned command surface

### Tasks

| Command | Purpose |
| --- | --- |
| `taskgo add <title>` | Capture a Task; `-d/--description`, `--priority`, `-p/--project`, `--inbox` |
| `taskgo list` | List Tasks; `--state=open\|done\|all`, `--visibility=active\|archived\|all`, `--all`, `-p/--project`, `--here`, `--priority` |
| `taskgo next` | Recommend the most important open Task; `--count N`, `-p/--project`, `--here` |
| `taskgo show <id>` | Show one Task in full |
| `taskgo done <id>` | Mark a Task Done |
| `taskgo reopen <id>` | Reopen a Done Task |
| `taskgo edit <id>` | Change title, description, priority, or Project |
| `taskgo archive <id>` | Hide a Task without changing completion |
| `taskgo unarchive <id>` | Restore an archived Task |
| `taskgo delete <id>` | Permanently delete a Task; `--force` skips confirmation |

### Projects

| Command | Purpose |
| --- | --- |
| `taskgo project add <name>` | Create a Project; `--repo <path>` binds a repository |
| `taskgo project list` | List Projects; `--visibility=active\|archived\|all` |
| `taskgo project rename <name> <new-name>` | Rename a Project |
| `taskgo project edit <name>` | Rebind with `--repo`, clear with `--clear-repo` |
| `taskgo project archive <name>` | Hide a Project and its Tasks |
| `taskgo project unarchive <name>` | Restore a Project and its Tasks |
| `taskgo project delete <name>` | Delete a Project; `--to-inbox` preserves Tasks, `--force` destroys them |

### Examples

```sh
taskgo add "fix flaky auth test" --priority high
taskgo add "call the dentist" --inbox
taskgo next --count 3
taskgo list --state done --json | jq '.[].title'
taskgo done 42
taskgo project add "API" --repo ~/work/api
```

## Output contract

- Human-readable by default; `--json` on every command.
- JSON lists are arrays, `show` returns one resource, `next` returns one resource or `null`, mutations return the affected resource.
- Data goes to stdout; errors and diagnostics go to stderr.
- Exit `0` success, `1` domain or runtime failure, `2` usage error. A declined confirmation exits `1`.
- JSON timestamps are RFC 3339 UTC; human output uses local time.
- Color is TTY-detected and disabled by `NO_COLOR` or `--no-color`. JSON never contains ANSI or diagnostics.
- JSON shapes are documented but explicitly **unstable until v1.0**.

## Data and storage

The database lives in the platform application-data directory:

| OS | Location |
| --- | --- |
| Linux | `$XDG_DATA_HOME/taskgo/taskgo.db`, else `~/.local/share/taskgo/taskgo.db` |
| macOS | `~/Library/Application Support/taskgo/taskgo.db` |
| Windows | `%LOCALAPPDATA%\taskgo\taskgo.db` |

`TASKGO_DB` is the only override. There is no `--db` flag and no config file.

- One database is shared by every Project, repository, and the Inbox.
- taskgo never creates repository-local metadata, so repositories stay clean.
- The database is portable as a file, but **synchronization is not supported**.
- SQLite runs in WAL mode. To back up safely, stop taskgo or use SQLite's backup API; copying a live database can miss committed transactions.
- Network filesystems and synced folders (NFS, SMB, Dropbox, etc.) are unsupported.

## v1 scope

Included: Tasks, Projects with an optional repository path, the Inbox, priority, completion, archiving, deletion, detection, JSON output, and migrations.

Explicitly out: remote sync, multi-device or multi-user use, remote trackers, import/export, subtasks, dependencies, tags, due dates, reminders, recurrence, estimates, a TUI, a daemon, an HTTP API, plugins, per-repository databases, custom workflows, text search, custom sorting, editors, batch capture, and statistics.

## Documentation

| Document | Contents |
| --- | --- |
| [`GLOSSARY.md`](GLOSSARY.md) | The vocabulary: Task, Project, Inbox, Archived, Detected Project, and more |
| [`docs/adr/`](docs/adr/) | Six architectural decisions, including the domain model, storage, IDs, and detection |
| [`.scratch/taskgo-v1/spec.md`](.scratch/taskgo-v1/spec.md) | The full v1 spec and testing seams |
| [`.scratch/taskgo-v1/issues/`](.scratch/taskgo-v1/issues/) | The 12 implementation tickets |
| [`AGENTS.md`](AGENTS.md) | Working rules for agents in this repository |

## Development

```sh
make help     # list targets
make build    # build bin/taskgo
make test     # run tests
make check    # format, vet, and test
make clean    # remove build artifacts
```

Architecture: CLI → application → domain, with SQLite behind an application port. Command handlers parse flags and render output; they contain no business rules.
