---
status: accepted
---

# Tasks live in their own repository

Task files move from the code repository to their own `fuda-` tasks repository. fuda reads and writes its default branch only. The code repository mounts it as a git submodule at the old path (`docs/board`), with `branch = main`, and nobody commits the pointer: people and agents run `git submodule update --remote docs/board`. We did this so there is one branch and one Status per Task, and so people who need only Tasks get access to Tasks only.

## Considered Options

- A board branch in the code repository that fuda writes to: rejected. Agents and the board would see two Statuses (develop and the board branch), and it needs a merge back to develop.
- A bot or hook that bumps the submodule pointer: rejected. Noise commits and merge conflicts.

## Consequences

- A code change and its Task edit are two commits in two repositories.
- Code PRs no longer change Task files, so In review can no longer be matched by the PR's version of the Task file. A PR names its Task by id in its title or branch name instead.
- Existing repositories migrate with `git subtree split --prefix=docs/board`, which keeps the history.
