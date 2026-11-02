---
status: accepted
---

# Binary task lifecycle; archiving is visibility

A Task is either open or done, represented solely by whether its completion time is set. There is no in-progress state, no configurable workflow, and no cancelled state. Archiving is a separate, reversible visibility flag that applies to open and done Tasks alike, so abandoned work has an honest home without pretending to be finished.

## Considered options

- **Three states including in-progress.** Familiar, but the state rots: tasks stay "in progress" for weeks and the label stops carrying information.
- **Per-project workflows.** Planning-engine territory, excluded from v1.
- **Cancelled as a workflow state.** Confuses outcome with visibility; archiving already covers abandonment.

## Consequences

- "What am I working on?" is answered by selection (`next`), not by a stored state. A start timestamp can be added later without changing the lifecycle.
- A future board view needs a deliberate state-machine design rather than inheriting one.
- Archiving an open Task never completes it, and repeated completion never overwrites an existing completion time.
