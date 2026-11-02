# AGENTS.md

## Git

- **Never** stage, commit, push, create branches, tag, or rewrite history without the author's explicit permission.
- Always ask permission first, every time. Permission for one commit is not permission for the next.
- When work is ready, propose a branch name and/or commit message for the author to run themselves.
- Never commit on the author's behalf, even when asked to "implement" or "finish" a task.
- Read-only Git commands (`git status`, `git diff`, `git log`) are fine without asking.

## Agent skills

### Issue tracker

Issues live as markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical roles, each label string equal to its role name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` plus `docs/adr/` at the repo root. See `docs/agents/domain.md`.
