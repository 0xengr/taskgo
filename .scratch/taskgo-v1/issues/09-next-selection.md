# 09: `next` selection

**What to build:** `taskgo next` recommends the single most important open Task across every Project and the Inbox, using the canonical order, with optional widening and scoping. It never mutates state or remembers a cursor.

**Blocked by:** 03

**Status:** ready-for-agent

- [ ] `next` considers all open, unarchived Tasks whose Project is not archived.
- [ ] Ordering is `priority DESC, created_at ASC, id ASC`; the top row of `list` matches `next`.
- [ ] It returns one Task by default.
- [ ] `--count N` returns the top N from the same ordering; zero or negative is a usage error (exit 2).
- [ ] `--project` scopes to one Project; `--here` scopes to the Detected Project.
- [ ] `--here` outside an active Project is a domain error (exit 1), including when the deepest match is archived.
- [ ] With nothing available, the command exits 0, prints a human message, and emits `null` with `--json`.
- [ ] Repeated invocations return the same result until the data changes.
