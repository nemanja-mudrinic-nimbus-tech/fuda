# Which branch holds a Task's current Status

Type: grilling
Status: resolved
Blocked by: 01, 04

## Question

With Moves and Assigns on a board branch and developers still editing Task files on develop, which branch does the board read a Task's Status and Owner from? What does the board show when the two disagree, and what do agents read when they open a Task file on develop?

## Answer

Decided with the user, 2026-10-08. This replaces the board branch.

- Task files move to their **own tasks repository**. Fuda reads and writes its default branch. There is one branch, so the board and agents never see two Statuses.
- The code repository mounts the tasks repository as a **git submodule** at the old path (for example `docs/board`). Fuda does not use the submodule; it talks to the tasks repository directly.
- People who need only Tasks get access to the tasks repository only.
- Cost: the submodule pointer in the code repository gets old. Developers and agents must pull the newest Tasks. How, and where that is explained, is the ticket "How developers and agents keep the tasks submodule up to date".
- A code change and its Task edit are now two commits in two repositories.
- ADR needed: hard to reverse, surprising, a real trade-off (rejected: a board branch in the code repository).
