# 04: Task detail and completion

**What to build:** Developers can inspect one Task in full, mark it Done, and Reopen it. Completion is idempotent and never rewrites an existing completion time.

**Blocked by:** 03

**Status:** ready-for-agent

- [ ] `show` renders the full Task: title, description, priority, Project or Inbox, completion state, archive state, and timestamps.
- [ ] `done` sets the completion time only when the Task is open.
- [ ] Repeating `done` is a successful no-op that preserves the original completion time and returns the unchanged resource.
- [ ] `reopen` clears completion; repeating it on an open Task is a no-op.
- [ ] Archived Tasks remain addressable by ID.
- [ ] An unknown ID is a domain error (exit 1).
- [ ] `--json` returns one resource with RFC 3339 UTC timestamps; human output formats times locally.
