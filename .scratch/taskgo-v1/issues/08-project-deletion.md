# 08: Project deletion

**What to build:** Deleting a Project never silently destroys work. A Project with Tasks requires an explicit strategy: move its Tasks to the Inbox, or confirm permanent destruction of the Project and its Tasks.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] `project delete` on a Project with Tasks and no strategy refuses, explains the available options, and exits 1.
- [ ] `--to-inbox` reassigns every Task to the Inbox by clearing its Project association, preserving IDs and all other fields, in one transaction.
- [ ] `--force` selects permanent deletion of the Project and its Tasks but still requires confirmation.
- [ ] Declining confirmation exits 1 and changes nothing.
- [ ] Deleting an empty Project works with confirmation.
- [ ] In a non-TTY session, `--force` fails with an actionable message rather than hanging.
- [ ] An unknown Project name is a domain error (exit 1).
- [ ] `--json` reports the Project and, for `--to-inbox`, the number of preserved Tasks.
