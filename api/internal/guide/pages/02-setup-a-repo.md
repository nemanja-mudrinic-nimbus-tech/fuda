# Set up a repo

A board is the `docs/board/` folder on `main` of a repository, and you choose where that repository
is. It can be a tickets repository of its own, named `fuda-<project>` (for example `fuda-eliot`;
fuda lists every `fuda-` repository the person can read), or a code repository that you list in
`FUDA_GITHUB_BOARDS` or `FUDA_AZURE_BOARDS`, with the tasks in its own `docs/board/`. fuda reads
and writes `main`, the only branch it uses, in both. Code repositories listed in `repos.md` put
their pull requests' tasks In review.

## What lives where

| Where | What | Read by |
|---|---|---|
| Board repo: task files (required) | `docs/board/tasks/*.md`, `docs/board/archive/*.md` | fuda and people |
| Board repo: board config (optional) | `docs/board/stages.md`, `labels.md`, `people.md`, `repos.md` | fuda and people |
| Board repo: instructions | `TASKS.md`, `WORKFLOW.md`, the agent guide | people and agents, never fuda |
| Code repos with their own tickets repository: pointer | The agent guide's pointer in `CLAUDE.md` / `AGENTS.md` | agents, never fuda |
| fuda's environment | GitHub App client id | fuda only |
| GitHub | who may read the repository: each person logs in with their own account | GitHub |

What the board looks like belongs to your repository. How fuda reaches it belongs to fuda's
environment.

## Folder layout

```
<board-repo>/                  a `fuda-<project>` repository, or a code repository you list; branch main
  docs/                        every markdown file here opens in the Docs reader and shows in its file tree
                               (except tasks, the archive and the board config files below)
    board/
      tasks/                   required: one file per task, flat
      archive/                 tasks off the board, any status (the Archive view)
      stages.md                optional: the columns
      labels.md                optional: label groups, values and colours
      people.md                optional: people, ids and agents
      repos.md                 optional: code repositories whose PRs put tasks In review
      TASKS.md                 recommended: the workflow, copied from this guide
      WORKFLOW.md              optional: your own rules; fuda ignores it
  .claude/skills/fuda-tasks/   the agent guide
```

Tasks sit directly in `tasks/`, or one level down in an epic folder: `tasks/<epic>/T-1.md`. The
folder name is the task's epic (it gets the label `epic:<folder>`), so moving a task to another epic
is a `git mv`. Deeper folders are listed under Problems. `archive/` follows the same layout.
Ids are unique across all folders. fuda reads and writes `main` of the board repository. It must be
able to push there directly: if `main` is protected, a move fails with a message saying so. In a
tickets repository, allow fuda's users to push to `main` (a bypass rule or a ruleset exception). In a
code repository, the message tells them to change the task in a pull request instead.

The In review column still comes from open pull requests in the code repositories listed in
`repos.md`.

## Checklist

1. Create `docs/board/tasks/` and add or migrate tasks ([Task format](/guide/task-format)). An
   agent with the [agent guide](/guide/agent-guide) can migrate an old backlog.
2. Optionally add `stages.md`, `labels.md`, `people.md` and `repos.md` ([Board config files](/guide/board-config)).
3. Copy the workflow into `TASKS.md`. Add the agent guide to the board repository. When the board
   is a tickets repository, add the `CLAUDE.md` pointer to each code repository.
4. Optionally make the code repositories' CI skip pull requests that only touch `docs/board/` ([Workflow](/guide/workflow#ci-cd)).
5. Try it locally ([Run locally](/guide/run-locally)) and fix anything listed under Problems.
6. Push to `main`. For a tickets repository, the name starts with `fuda-`; for a code repository, list
   it in `FUDA_GITHUB_BOARDS` or `FUDA_AZURE_BOARDS`. Install the fuda GitHub App on the organisation
   or user that owns it, and deploy fuda ([Self-host](/guide/self-host)). Everyone who can read the
   repository sees it as a Board.

A board with no task files is not an error. It shows an empty board with the default stages
and the hint "No Tasks yet: add files in `tasks/`". One with no `main` branch shows an error page that
says how to create it.

## Move an old `develop` board

Boards that kept tasks on the code repository's `develop` branch are no longer read: fuda reads
`main` only. Either merge `docs/board` into `main` of the code repository and list that repository
in `FUDA_GITHUB_BOARDS` (or `FUDA_AZURE_BOARDS`), or move the tasks into their own repository and
keep their history:

1. Create an empty repository named `fuda-<project>` and install the fuda GitHub App on it.
2. In the code repository, on the branch that holds the tasks, split `docs/board` out:
   `git subtree split --prefix=docs/board -b board-split`.
3. Push it as `main`: `git push git@github.com:<owner>/fuda-<project>.git board-split:main`.
4. In the code repository, remove the folder: `git rm -r docs/board`, and commit. The code
   repository keeps no copy of the tasks.
5. List the code repository in the new board's `repos.md`, and delete the `board-split` branch.

The steps with explanations are in CONTRIBUTING.md of the fuda repository.
