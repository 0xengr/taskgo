---
status: accepted
---

# Repository detection uses canonical paths, with an archived barrier

The Project for a capture is found by walking the canonical current directory upward and matching registered repository paths by component-aware containment; the deepest match wins. Archive status does not participate in that search. Once the deepest matching Project is found, an archived match routes the capture to the Inbox instead of falling back to an enclosing active Project.

## Considered options

- **Git-root detection.** Fails inside submodules and worktrees, and does nothing for non-Git directories.
- **Longest prefix over raw strings.** A repository at `/work/api` would incorrectly capture `/work/api2`.
- **Filter archived Projects before matching.** Silently re-homes work into a parent Project when the user is inside an archived workspace.

## Consequences

- Nested Project paths are supported and deterministic; the most specific matching Project wins, and its archive state then decides between that Project and the Inbox.
- Worktrees and moved directories stop matching until re-bound with `project edit --repo`; there is no automatic repair.
- `--here` reports an error when the deepest match is archived, rather than pretending no Project exists.
- Symlinks are resolved at registration and lookup; bind-mount ancestry is not traversed.
