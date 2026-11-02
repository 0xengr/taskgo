---
status: accepted
---

# Task IDs are global, monotonic, and never reused

Tasks are referenced by a single integer sequence shared by all Projects and the Inbox, displayed as `#42`. The promise is stronger than uniqueness: an ID is never reused after archiving or deletion, so an ID written into a commit message, note, or shell history can never later point at different work.

## Considered options

- **Per-project numbering (`API-12`).** Readable, but requires a scope and counter per Project plus a special Inbox scope.
- **UUID or ULID.** Merge-safe but unusable to type in a terminal; a short alias would be added anyway.
- **Title-based lookup.** Ambiguous with duplicate titles and renames.

## Consequences

- IDs are local-database identifiers, not machine-independent ones. Synchronization or database merging would require an added UUID column and merge key.
- SQLite must use `INTEGER PRIMARY KEY AUTOINCREMENT`; plain rowids recycle values and would break the promise.
- Deleted IDs leave permanent gaps.
