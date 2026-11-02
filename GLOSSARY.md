# taskgo

Taskgo is a local-first command-line tracker for developers working in a terminal. It organizes work into Projects and Tasks, with an Inbox for work that has not been assigned yet.

## Language

### Work

**Task**:
A single unit of work that can be captured, prioritized, completed, archived, or deleted.
_Avoid_: Ticket, issue, item, todo

**Project**:
A logical workspace or initiative that groups Tasks and may be bound to one local repository path.
_Avoid_: Repo, board, epic, workspace (as a state)

**Inbox**:
The set of Tasks that are not assigned to any Project. It is not a Project; a Task leaves the Inbox when it is assigned.
_Avoid_: Default project, unassigned project

**Priority**:
An ordinal classification of importance — low, medium, or high. Not a score, and not a deadline.
_Avoid_: Severity, urgency, rank

### Lifecycle

**Open**:
A Task whose work is not finished.
_Avoid_: Todo, pending, active, in progress

**Done**:
A Task whose work is finished, recorded with the time it was completed.
_Avoid_: Closed, completed, finished

**Archived**:
A Task or Project hidden from default views without changing whether its work is finished. Archiving is reversible and is not a workflow stage.
_Avoid_: Deleted, cancelled, closed

**Deleted**:
Removed permanently and unrecoverably. Distinct from Archived.
_Avoid_: Removed, purged

### Capture

**Capture**:
Recording a Task with minimal ceremony, before any decision about where it belongs or how important it is.
_Avoid_: Create, add, log

**Triage**:
The deliberate act of assigning or prioritizing captured work.
_Avoid_: Sort, organize, groom

**Repository path**:
The local directory a Project is bound to, used to recognize that work is happening inside that Project's workspace. Not an independent thing in its own right.
_Avoid_: Repo record, repository link

**Detected Project**:
The Project inferred from the current directory at capture time. The deepest matching Project wins.
_Avoid_: Current project, default project

**Archived barrier**:
The rule that a Detected Project which is archived sends capture to the Inbox rather than falling back to an enclosing Project.
_Avoid_: Fallback, parent project

### Identity

**Task ID**:
The stable, never-reused reference to a Task.
_Avoid_: UUID, key, ticket number

**Project name**:
The user-facing identifier of a Project; unique regardless of letter case.
_Avoid_: Slug, title
