---
status: accepted
---

# Project owns a repository path; no Repository entity

Capturing tasks near a repository tempts a first-class Repository that Tasks link to. We deliberately did not add one: a Project owns at most one optional local repository path, and every Task belongs to exactly one Project or to the Inbox. No v1 behavior — auto-assignment, filtering, git context — needs repository identity, and nothing required a many-to-many link, so the cheaper model wins.

## Considered options

- **Repository entity with task↔repository links.** Supports cross-repository tasks and per-repository settings, but requires identity (clones, worktrees, remotes), lifecycle, and a join table before any feature needs them.
- **Repository as a Task attribute instead of a Project attribute.** Keeps multi-repository Projects, but loses deterministic directory-to-Project detection, which is the reason the binding exists.
- **No repository association at all.** Makes Projects indistinct from tags.

## Consequences

- Cross-repository work is represented as separate Projects and is outside v1 scope.
- One canonical repository path belongs to at most one Project.
- Adding a Repository entity later is additive; removing one after data exists is not.
