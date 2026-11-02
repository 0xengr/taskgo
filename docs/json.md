# JSON output

Every command accepts `--json`. Data goes to **stdout**; errors and diagnostics
stay on stderr, and the exit code carries the machine-readable outcome.

> The shapes below are **unstable until v1.0**. Fields may be added or renamed
> in a minor release; see the release notes.

## Conventions

- Timestamps are RFC 3339 in UTC.
- Absent values are `null`, not omitted.
- `id` is a number for Tasks and Projects.
- Output never contains ANSI escapes or diagnostic text.

## Task

```json
{
  "id": 1,
  "title": "fix flaky auth test",
  "description": null,
  "priority": "high",
  "project": { "id": 1, "name": "API" },
  "created_at": "2026-11-02T09:47:00Z",
  "completed_at": null,
  "archived_at": null
}
```

`project` is `null` for Tasks in the Inbox. `description` is `null` when absent.
`priority` is `low`, `medium`, or `high`.

## Project

```json
{
  "id": 1,
  "name": "API",
  "repository_path": "/home/me/work/api",
  "created_at": "2026-11-02T09:40:00Z",
  "archived_at": null
}
```

`repository_path` is `null` when the Project is not bound to a repository.

## Command shapes

| Command | Shape |
| --- | --- |
| `add`, `show`, `done`, `reopen`, `archive`, `unarchive`, `edit` | one Task |
| `delete` | one Task, as it existed immediately before deletion |
| `list` | an array of Tasks (possibly empty) |
| `next` (default count) | one Task, or `null` when nothing is open |
| `next --count N` (N > 1) | an array of up to N Tasks |
| `project add`, `rename`, `edit`, `archive`, `unarchive` | one Project |
| `project list` | an array of Projects (possibly empty) |
| `project delete --to-inbox` | `{ "project": Project, "moved_tasks": number }` |
| `project delete --force` | `{ "project": Project }` |
| `version` | `{ "version": string, "commit": string }` |

## Errors

Errors are human-readable text on stderr, even with `--json`; stdout stays
empty. Scripts should branch on the exit code:

| Exit | Meaning |
| --- | --- |
| `0` | success |
| `1` | domain or runtime failure (not found, conflict, busy, corrupt database, declined confirmation) |
| `2` | usage error (unknown flag, missing argument, invalid value, conflicting flags) |
