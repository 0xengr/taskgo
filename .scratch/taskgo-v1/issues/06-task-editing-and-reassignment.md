# 06: Task editing and reassignment

**What to build:** Developers can change a Task's title, description, priority, and Project (or move it to the Inbox) in one transactional operation, with clear usage and domain errors.

**Blocked by:** 02, 04

**Status:** ready-for-agent

- [ ] `edit` applies all supplied fields in a single transaction: title, description, priority, project, inbox.
- [ ] An empty description clears the description; an omitted field is left unchanged.
- [ ] `--project` and `--inbox` together are a usage error (exit 2).
- [ ] `edit` with no fields supplied is a usage error (exit 2).
- [ ] Assigning a Task into an archived Project is a domain error (exit 1).
- [ ] Unknown Task or Project names are domain errors (exit 1).
- [ ] `--json` returns the affected Task; human output confirms the change without dumping the whole record.
- [ ] Archived Tasks remain editable.
