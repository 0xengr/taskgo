# 01: Walking skeleton — capture and list

**What to build:** From a clean machine, `taskgo add "title"` records a Task in the Inbox and `taskgo list` shows it. The database is created automatically at the platform data directory (`TASKGO_DB` overrides it), schema migration 0001 applies transactionally, and both commands support human and `--json` output with the agreed exit codes.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] First invocation creates the database and schema; later invocations reuse it.
- [ ] `add` accepts a title, optional description, and priority; creates an open, unarchived Task in the Inbox and prints its `#ID` and destination.
- [ ] Title validation: trimmed, single-line, whitespace-normalized, at most 500 runes after normalization, control characters rejected, empty rejected (exit 2).
- [ ] Invalid priority and conflicting input are usage errors (exit 2).
- [ ] `list` defaults to open, unarchived Tasks ordered `priority DESC, created_at ASC, id ASC`; an empty result exits 0.
- [ ] `--json` emits Task resources with RFC 3339 UTC timestamps and no ANSI; human output uses local times.
- [ ] A process-boundary test helper builds the binary once, runs it with an isolated `TASKGO_DB`, and asserts stdout, stderr, and exit code.
- [ ] Migration from an empty database is transactional and leaves `user_version` at 0001.
