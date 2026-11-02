# 05: Task archiving and list dimensions

**What to build:** Tasks can be archived and unarchived as a visibility operation independent of completion, and `list` exposes completion and archive as two independent dimensions so every combination is expressible.

**Blocked by:** 02, 04

**Status:** ready-for-agent

- [ ] `archive` sets the archive time only when unarchived; `unarchive` clears it; both are idempotent.
- [ ] Archiving an open Task does not mark it Done, and unarchiving restores its previous open/Done state.
- [ ] `list` supports `--state=open|done|all` defaulting to `open`.
- [ ] `list` supports `--visibility=active|archived|all` defaulting to `active`.
- [ ] Archived visibility includes both Tasks archived directly and Tasks belonging to archived Projects.
- [ ] `--all` is accepted as a shortcut for both dimensions being `all`.
- [ ] Combining `--all` with an explicit state or visibility is a usage error (exit 2), as is an invalid enum value.
- [ ] Archived Tasks are excluded from the default view and remain reachable with `show`.
