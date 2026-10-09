# fuda

A kanban board and doc reader over a repository's markdown task files.

## Language

**Board**:
One tasks repository, shown as Stages and Tasks. One fuda can show several Boards.

**Local Board**:
A Board read from a folder on disk instead of from a git host. A Move or an Assign changes the file on disk; the person commits it.

**Task**:
One unit of work, stored as one markdown file in the repository.
_Avoid_: ticket, issue, card (a card is only how a Task is drawn on the board)

**Status**:
The field on a Task that says where it is in the workflow.

**Stage**:
A column on the board. One Stage groups one or more Statuses.
_Avoid_: lane, state

**Owner**:
The person or people working on a Task.
_Avoid_: assignee

**Move**:
Changing a Task's Status from fuda, so it appears in another Stage.

**Assign**:
Changing a Task's Owner from fuda.

**Problem**:
A repository file fuda could not read. It is shown to people, never fatal.

## Relationships

- A **Stage** groups one or more **Statuses**; a **Task** has exactly one **Status**.
- A **Move** and an **Assign** are the only changes fuda makes to a **Task**. Creating and editing Task text happens in git.

## Flagged ambiguities

- "ticket" and "issue" were used for **Task**. Resolved: the term is **Task**.
- "assign" was used for setting the **Owner**. fuda's file field is `owner`; the action is **Assign**.
