# 10: Reliability hardening

**What to build:** The database layer fails safely and predictably: newer schemas are refused, corrupt files are never recreated, lock contention produces an actionable error, concurrent processes stay consistent, and failed migrations leave the last good version intact.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] A database whose schema version is newer than the binary is refused with a clear message and exit 1.
- [ ] A corrupt or unreadable database produces an actionable diagnostic and is never deleted or recreated automatically.
- [ ] A write lock held beyond the busy timeout produces an actionable error and exit 1 rather than hanging.
- [ ] Concurrent captures from multiple processes produce unique IDs and no corruption.
- [ ] A deliberately failing migration rolls back completely, leaving the previous schema version intact.
- [ ] Schema version is read under `BEGIN IMMEDIATE` so two processes starting simultaneously cannot apply the same migration twice.
- [ ] The upgrade path from a version 0001 fixture is covered alongside fresh install and failure cases.
- [ ] Lock, corruption, and migration failures are exercised through the process-boundary seam.
