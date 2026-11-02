Status: ready-for-agent

# taskgo v1

## Problem Statement

Developers accumulate work faster than they can track it, and the work is scattered across repositories, terminals, and context switches. Existing tools force a choice: heavyweight cloud trackers that are slow to capture into and unavailable offline, or generic todo apps that know nothing about repositories and can't be scripted usefully. A thought captured mid-task either gets lost or interrupts the work to file it properly.

## Solution

`taskgo` is a local-first command-line tracker for developers. It stores everything in one SQLite database on the local machine, captures a Task in a single command without requiring a Project decision, and uses the current directory to infer the Project when the developer is working inside a registered Repository path. Triage is a deliberate, separate act. Every command supports human-readable output by default and `--json` for scripting, with a strict exit-code contract.

## User Stories

**Capture**

1. As a developer, I want to capture a Task with only a title, so that I can record a thought without breaking flow.
2. As a developer working inside a registered repository, I want the Task assigned to that Project automatically, so that I don't have to name it every time.
3. As a developer, I want the destination printed after capture, so that I can trust where the Task went.
4. As a developer capturing an unrelated thought while inside a repository, I want `--inbox` to override detection, so that work and personal notes don't mix.
5. As a developer, I want `--project` to target a specific Project explicitly, so that I can capture for another workspace.
6. As a developer, I want to attach a multiline description, so that a Task can carry enough context to act on later.
7. As a developer, I want to set priority at capture time, so that urgent work is distinguishable immediately.
8. As a developer, I want duplicate titles to be allowed, so that recurring work isn't silently merged.
9. As a developer pasting messy text, I want titles trimmed and whitespace-normalized, so that the list stays readable.
10. As a developer, I want an over-long or control-character title rejected with a usage error, so that bad data never enters the database.
11. As a developer, I want an unknown Project name to fail as a domain error, so that scripts can distinguish it from a typo in my flags.
12. As a scripter, I want `--json` capture to return the created Task, so that I can chain the new ID.

**Detection**

13. As a monorepo developer, I want nested Projects to resolve to the most specific one, so that captures land where I'm actually working.
14. As a developer, I want captures inside an archived Project's path to go to the Inbox rather than a parent Project, so that finishing an initiative doesn't silently redirect new work.
15. As a developer using symlinked directories, I want detection to work through the symlink, so that my working setup doesn't change behavior.
16. As a developer who moved a repository, I want detection to stop matching quietly, so that nothing is auto-rebound behind my back.
17. As a developer outside any registered repository, I want capture to go to the Inbox, so that capture never fails for lack of context.

**Browsing and selection**

18. As a developer, I want `list` to show open, unarchived Tasks by default, so that the daily view is the work that matters.
19. As a developer, I want `--state=done` to inspect finished work, so that I can review what I completed.
20. As a developer, I want `--state=all` to see everything without archive noise, so that I can audit my backlog.
21. As a developer, I want `--visibility=archived` to include Tasks archived directly and Tasks belonging to archived Projects, so that one flag covers both ways work disappears.
22. As a developer, I want `--all` as a shortcut for both dimensions, so that "show me everything" is one flag.
23. As a developer, I want combining `--all` with an explicit dimension to be a usage error, so that the result is never ambiguous.
24. As a developer, I want `--here` to scope to my detected Project, so that I can ignore other workspaces.
25. As a developer running `--here` outside a Project, I want a clear error, so that I don't mistake an empty list for "no work."
26. As a developer, I want to filter by Project or priority, so that I can focus a review.
27. As a developer, I want a stable canonical order, so that the top row of `list` matches what `next` recommends.
28. As a developer, I want `next` to consider all open work across Projects and the Inbox, so that I get the most important thing regardless of where it lives.
29. As a developer with a blocked top Task, I want `--count N` to widen the same ordered list, so that I can choose without a rotation cursor.
30. As a developer, I want `next --here` to scope to my current Project, so that I can ask "what next here?"
31. As a developer with nothing open, I want `next` to succeed with an empty result, so that a status-line poller doesn't treat a clear queue as an error.
32. As a developer, I want `show` to display one Task in full, so that I can read the description without listing everything.

**Lifecycle**

33. As a developer, I want `done` to record completion time, so that throughput is observable.
34. As a developer, I want repeated `done` calls to be no-ops that preserve the original completion time, so that re-running a script never rewrites history.
35. As a developer who completed something by mistake, I want `reopen` to clear completion, so that I can correct it.
36. As a developer abandoning work, I want to archive an open Task without marking it done, so that my completion count stays honest.
37. As a developer, I want to unarchive and get the Task's original state back, so that archiving is safely reversible.
38. As a developer, I want to edit title, description, priority, and Project together in one operation, so that fixes are atomic.
39. As a developer, I want an empty description flag to clear the description, so that removal doesn't need a second command.
40. As a developer, I want moving a Task into an archived Project to fail, so that I never create work that immediately vanishes.
41. As a developer, I want `edit` with no fields to be a usage error, so that I know nothing happened.
42. As a developer, I want Task deletion to prompt by default, so that a mistyped ID doesn't destroy history.
43. As a scripter, I want `--force` on Task deletion to skip the prompt, so that cleanup can be automated.
44. As a developer, I want deleting an unknown or already-deleted Task to fail, so that success always means something happened.

**Projects**

45. As a developer, I want to create a Project bound to a local repository, so that captures can be detected.
46. As a developer, I want Project names compared case-insensitively, so that `api` and `API` aren't two Projects.
47. As a developer, I want to rename a Project without breaking its Tasks, so that early naming mistakes are cheap.
48. As a developer, I want to rebind or clear a Project's repository path, so that moved directories and non-repo Projects are handled.
49. As a developer, I want registering a path that doesn't exist to fail, so that typos don't create dead bindings.
50. As a developer, I want archiving a Project to hide its Tasks without changing them, so that finishing an initiative doesn't rewrite history.
51. As a developer, I want unarchiving a Project to restore exactly what was hidden, so that no Task is silently unarchived.
52. As a developer, I want `project list` to default to active Projects and to reveal archived ones on request, so that old Projects stay discoverable.
53. As a developer deleting a Project, I want the command to refuse while Tasks exist, so that destruction is never accidental.
54. As a developer, I want `--to-inbox` to preserve a deleted Project's Tasks, so that cleanup doesn't lose work.
55. As a developer, I want `--force` on Project deletion to still require confirmation, so that the destructive path is deliberate.

**Interface and reliability**

56. As a scripter, I want data on stdout and diagnostics on stderr, so that pipes stay clean.
57. As a scripter, I want `0/1/2` exit codes, so that I can distinguish success, domain failure, and bad usage.
58. As a scripter, I want JSON timestamps in RFC 3339 UTC, so that values parse without locale assumptions.
59. As a developer, I want local human timestamps and TTY-aware color that respects `NO_COLOR`, so that output is readable without polluting pipes.
60. As a new user, I want the database created automatically on first run, so that there is no setup step.
61. As a user with two terminals open, I want concurrent reads and writes to behave, so that a status-line poller doesn't break a capture.
62. As a user with a newer database, I want a refusal with a clear explanation, so that an old binary never corrupts new data.
63. As a user with a corrupt or unreadable database, I want an actionable error and no automatic recreation, so that recovery is my decision.
64. As a user upgrading taskgo, I want migrations to be transactional and forward-only, so that a failure leaves the last good schema intact.
65. As a user, I want `version` to report the build, so that bug reports are actionable.

## Implementation Decisions

**Layers.** CLI → application → domain, with SQLite behind an application port. Command handlers parse flags, call one application operation, render output, and return an exit code. All invariants live in domain/app; handlers contain none. No package outside infrastructure imports SQLite.

**Schema.**
- `projects`: internal ID, display `name`, `name_key` (unique, case-folded), nullable `repo_path` (unique when present), `created_at`, nullable `archived_at`.
- `tasks`: `INTEGER PRIMARY KEY AUTOINCREMENT` ID, `title`, nullable `description`, `priority`, nullable `project_id`, `created_at`, nullable `completed_at`, nullable `archived_at`.
- Indexes support the canonical order and the default filters (`project_id`, completion, archive).
- `PRAGMA user_version` is the schema version. `foreign_keys=ON`, WAL, `busy_timeout=5000`, `synchronous=NORMAL`.

**Identity.** Task IDs are global and never reused; deletion leaves gaps. Human output prefixes `#`; JSON emits the integer. Project identity is the name via `name_key = fold(NFC(trim(name)))` computed in Go with Unicode full folding; SQLite `COLLATE NOCASE` is never used (ADR-0006).

**Capture.** Title required, trimmed, single-line, whitespace-run-collapsed with Unicode-aware whitespace, ≤500 runes after normalization, controls rejected. Description optional and multiline; empty means absent. Priority is `low|medium|high`, case-insensitive, default `medium`. `--project` and `--inbox` conflict. One Task per invocation. Unknown Project is a domain error; malformed input is a usage error.

**Detection.** Resolve the canonical current directory; find the deepest registered Repository path containing it by component-aware comparison; if that Project is archived, resolve to Inbox and stop (no parent fallback). No match → Inbox. Paths are canonicalized and symlink-resolved at registration; missing paths simply don't match. Nested registrations are allowed.

**Selection.** Canonical order is `priority DESC, created_at ASC, id ASC` for both `list` and `next`. `list` exposes `--state=open|done|all` and `--visibility=active|archived|all`, defaulting `open` + `active`; archive visibility is the union of Task-archived and Project-archived. `--all` aliases both, and conflicts with explicit dimensions. `next` is always open + active, one result, `--count N`, `--project`, `--here`. No cursor, search, sort, or limit.

**Mutations.** Idempotent setters; existing timestamps never overwritten. `edit` applies all supplied fields in one transaction. Archived Tasks remain addressable. Tasks cannot move into archived Projects. Task deletion prompts unless `--force`.

**Project lifecycle.** Archiving sets only the Project's archive time; unarchiving restores visibility without touching Tasks. Deleting with Tasks requires `--to-inbox` or `--force`; `--force` still confirms. Both paths transactional.

**Output.** Human default; `--json` everywhere. Lists → arrays, show → resource, next → resource or `null`, mutations → affected resource. Data on stdout; errors and diagnostics on stderr. Errors remain human-readable even with `--json`; the machine signal is the exit code. Exit `0` success, `1` domain/runtime failure, `2` usage error; declined confirmation `1`. RFC 3339 UTC in JSON, local in human output. Color is TTY-detected, disabled by `NO_COLOR`/`--no-color`, never in JSON. Schema is documented as unstable until v1.0.

**Storage.** One global database in the OS application-data directory; `TASKGO_DB` is the only override. No `--db`, no config file, no repository-local metadata. No forced WAL checkpoint as a correctness requirement.

**Migrations.** Numbered, embedded, forward-only, immutable once shipped; migration and `user_version` bump in one transaction; newer-than-binary schema refused; failures roll back completely; no automatic deletion or recreation.

**Commands.** Ten Task verbs (`add`, `list`, `next`, `show`, `done`, `reopen`, `edit`, `archive`, `unarchive`, `delete`), seven Project subcommands (`add`, `list`, `rename`, `edit`, `archive`, `unarchive`, `delete`), plus `version`. Help is provided throughout.

## Testing Decisions

**What makes a good test here.** Assert observable behavior — stdout, stderr, exit code, and the state revealed by a subsequent command — never internal calls or private functions. One behavior per test, table-driven where inputs vary. Tests must be independent and parallel-safe: every test gets its own `TASKGO_DB` under a temp directory.

**Primary seam — process boundary.** A single helper builds the binary once per test run, executes it with a controlled environment and working directory, and returns captured streams and status. It covers every command and every contract: capture and detection, list/next selection, mutation idempotency, JSON shapes, exit codes, color suppression, confirmation behavior (TTY vs non-TTY), migrations from zero, newer-schema refusal, corruption handling, and lock timeout.

**Secondary seam — pure domain functions.** Table and property tests for title normalization, project-name folding keys, priority ordering, and canonical-path containment. Unicode (NFC/NFD, full folding), whitespace, control characters, and path-boundary cases (`/work/api` vs `/work/api2`, `..`, symlinks) are exercised here because process-level property testing would be prohibitively slow.

**Migration matrix.** Fresh install from version zero; an upgrade from every supported prior schema version, using checked-in fixture databases; a deliberately failing migration that proves full rollback; a database whose version is newer than the binary.

**Concurrency and failure.** Parallel captures from multiple processes prove unique IDs and no corruption. A held write transaction proves the busy-timeout message and exit `1`. Read-only directories, unreadable files, and non-TTY deletion without `--force` prove actionable errors instead of hangs.

**Benchmarks.** A separate harness seeds 10,000 Tasks and measures p95 for `list`, `show`, `next`, and `project list`, including process startup and database open. Benchmarks do not run as part of `go test`, so timing noise never fails an integration test. The reference environment is documented alongside the numbers.

**Prior art.** None — the repository is greenfield. These are the first tests, so they also establish the conventions later code follows.

## Out of Scope

Remote synchronization, multi-device, and multi-user operation. Remote trackers and import/export. Subtasks, dependencies, tags, due dates, reminders, recurrence, estimates. TUI, daemon, HTTP API, plugins. Per-repository databases and repository-local metadata. Custom priorities and configurable workflows. Text search, `--limit`, and custom sorting. Editor integration, stdin capture, batch capture, and prompts. Automatic archiving, repair, or path rebinding. Statistics and reporting.

## Further Notes

Decisions derived during shaping but not separately ratified, and worth confirming at implementation time: the exact human table format for `list`; the `taskgo: ` error prefix; whether `--json` errors ever gain a structured envelope (currently: no, exit codes carry it); and the confirmation prompt wording.

Known risks carried into implementation: Unicode table upgrades can change stored `name_key` values; cross-platform path identity (Windows drives/UNC, macOS case-insensitivity) needs property coverage; two processes migrating simultaneously need a `BEGIN IMMEDIATE` guard; `project delete --force` is intentionally not scriptable because it always confirms.

Project-name case folding is recorded in ADR-0006. Because `name_key` is persisted, the folding version must stay pinned and any table upgrade requires a recompute migration — never a lazy recompute at read time.
