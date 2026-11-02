# 03: Detection and capture assignment

**What to build:** Capturing from inside a registered repository assigns the Task to that Project automatically. Explicit flags override detection, nested registrations resolve deterministically, and an archived Project's path sends capture to the Inbox instead of an enclosing Project.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] `add` inside an active Project's repository path assigns the Task to that Project and echoes the destination.
- [ ] With nested Project paths, the deepest matching Project wins.
- [ ] A capture inside an archived Project's path goes to the Inbox, never to an enclosing active Project.
- [ ] `--inbox` overrides detection; `--project` selects explicitly using the same name folding; using both together is a usage error (exit 2).
- [ ] An unknown Project name is a domain error (exit 1).
- [ ] A symlinked working directory matches the Project it points into.
- [ ] A moved or missing repository path stops matching without an error or automatic repair.
- [ ] Containment is component-aware: a Project at `/work/api` does not capture `/work/api2`.
- [ ] No match sends the Task to the Inbox.
