---
status: accepted
---

# Project names are unique under Unicode case folding, not SQLite NOCASE

Project names are user-facing identifiers, so `api` and `API` must not create two Projects. SQLite's `COLLATE NOCASE` only folds ASCII, which would make uniqueness silently script-dependent and wrong for most of the world's names. The unique key is therefore `name_key = fold(NFC(trim(name)))`, computed in Go with Unicode full case folding, stored beside the original name, and enforced by a unique index.

## Considered options

- **SQLite `COLLATE NOCASE`.** Zero dependencies, but ASCII-only. The failure is invisible until a non-ASCII name collides or fails to collide.
- **`strings.ToLower`.** Simple mapping without normalization or full folding; `STRASSE` and `straße` would diverge, and composed/decomposed forms would create duplicates.
- **Restrict names to ASCII.** Predictable, but rejects legitimate project names for no user-facing benefit.
- **No uniqueness.** Every command would need disambiguation, and `--project` would become ambiguous.

## Consequences

- Adds `golang.org/x/text` (`cases`, `norm`, `unicode/norm`). The dependency is small and widely used.
- Folding is locale-insensitive by design: no Turkish dotless-i special case, no locale-specific collation. Uniqueness is deterministic across machines.
- `ß` and `ss` collide, as full folding requires. Display always uses the original `name`; the key is never shown.
- Unicode table upgrades can change folded keys. Mitigation: pin the `x/text` version and, on upgrade, recompute keys in a migration that checks for collisions before adding the unique index back. A folded key must never be recomputed lazily at read time.
- Registration rejects names that are empty after trimming; the key is computed only from valid names.
