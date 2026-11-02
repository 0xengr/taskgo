# 07: Task deletion

**What to build:** A Task can be deleted permanently after an explicit confirmation, or without a prompt when `--force` is supplied for scripting. Deleted IDs are never reused.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] `delete` prompts before removing a Task.
- [ ] Confirming removes the Task permanently; declining exits 1 and leaves the Task untouched.
- [ ] `--force` skips the prompt and deletes immediately.
- [ ] In a non-TTY session without `--force`, the command exits 1 with an actionable message instead of hanging on stdin.
- [ ] Deleting an unknown or already-deleted Task is a domain error (exit 1).
- [ ] A deleted ID is never assigned to a new Task; the next Task receives a higher ID.
- [ ] `--json` returns the deleted Task resource as it existed immediately before deletion.
