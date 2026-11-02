---
status: accepted
---

# One global database; no repository-local storage

Taskgo stores everything in a single SQLite database in the operating system's application-data directory, overridable only by `TASKGO_DB`. Per-repository databases were rejected because global Task IDs, a single Inbox, and a global `next` all assume one ID space. Splitting storage would create disconnected inboxes and ID namespaces, and a global-plus-repository overlay is synchronization by another name.

## Considered options

- **Per-repository `.taskgo/` database.** Clean repository isolation, but incompatible with the global model and adds repository-local metadata.
- **Hybrid global database plus repository overlay.** Merge semantics without calling it synchronization.

## Consequences

- Copying the database file is the portability story; synchronization is not supported.
- WAL adds `-wal` and `-shm` sidecars, so a safe backup requires stopping taskgo or using SQLite's backup API.
- Network and synced filesystems are unsupported and documented as such.
- Repositories stay free of taskgo metadata and merge conflicts.
