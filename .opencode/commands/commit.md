---
description: Review pending changes and propose a commit (never commits without explicit permission)
---

Prepare a commit for the current repository. You are the author's assistant, not
the committer: the author must approve every write Git operation.

**Hard rules (from AGENTS.md)**
- Never push, tag, amend, rebase, reset, or force anything.
- Never create or switch branches on your own.
- Never stage or commit until the author explicitly says yes in this conversation.
- One approval covers one commit and one set of paths only.

Optional scope or intent from the author: $ARGUMENTS

## 1. Inspect (read-only)

Run and review:
- `git status --short --branch`
- `git diff --stat`
- `git diff` (unstaged)
- `git diff --cached` (already staged)
- `git log --oneline -10` (skip if the repo has no commits)

Identify the current branch and whether it is `main`/`master`/the default branch.
If the repository has no commits yet, say so and treat every tracked candidate as
a new file.

## 2. Review the change set

- Group the changes into one coherent commit. If they are clearly unrelated,
  propose splitting them and say which paths belong to which commit.
- Exclude generated artifacts, build output, editor noise, and anything real that
  `.gitignore` already covers (flag accidental additions instead of committing them).
- Flag risks before proposing anything: secrets or `.env` values, credentials,
  large or binary files, debug leftovers, commented-out code, and files that look
  unrelated to the stated intent.
- If the diff is too large to review confidently, say exactly which files you
  could not inspect.

## 3. Propose

Present a short plan containing:

1. **Branch** — the current branch. If it is the default branch, suggest a new
   branch name (e.g. `feat/short-slug`, `fix/short-slug`, `chore/short-slug`) and
   the command the author would run. Do not create it.
2. **Paths** — the exact files to stage, listed explicitly. Never plan a bare
   `git add -A` or `git add .`.
3. **Message** — a Conventional Commits message (`type(scope): summary`, imperative,
   under ~72 characters), plus a body only when the change genuinely needs
   explaining.
4. **Command** — the exact `git add <paths>` and `git commit -m "..."` commands,
   for transparency.

## 4. Ask

Use the `question` tool to ask whether to proceed, with the proposed message as
the recommended option. Also offer "adjust the message" and "I'll run it myself".

- If the author chooses to run it themselves, stop and hand over the commands.
- If the author asks for changes, revise and ask again.
- Only on an explicit yes, stage the listed paths and commit with the approved
  message.

## 5. Report

After a successful commit, show `git status --short --branch` and the new
`git log -1 --stat`. Do not push. Remind the author that pushing is their call.
