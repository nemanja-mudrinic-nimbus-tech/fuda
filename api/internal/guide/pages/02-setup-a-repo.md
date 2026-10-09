# Set up a repo

## What lives where

| Where | What | Read by |
|---|---|---|
| Your repo: task files (required) | `docs/board/tasks/*.md`, `docs/board/archive/*.md` | fuda and people |
| Your repo: board config (optional) | `docs/board/stages.md`, `labels.md`, `people.md`, `repos.md` | fuda and people |
| Your repo: instructions | `TASKS.md`, `WORKFLOW.md`, the agent guide, a pointer in `CLAUDE.md` / `AGENTS.md` | people and agents, never fuda |
| fuda's environment | GitHub App client id, whether to watch `main` | fuda only |
| GitHub | who may read the repository: each person logs in with their own account | GitHub |

What the board looks like belongs to your repository. How fuda reaches it belongs to fuda's
environment.

## Folder layout

```
<repo>/
  docs/                        every markdown file here opens in the Docs reader
    board/
      tasks/                   required: one file per task, flat
      archive/                 finished tasks (the Archive view)
      stages.md                optional: the columns
      labels.md                optional: label groups, values and colours
      people.md                optional: people and aliases
      repos.md                 optional: code repositories whose PRs put tasks In review
      TASKS.md                 recommended: the workflow, copied from this guide
      WORKFLOW.md              optional: your own rules; fuda ignores it
  CLAUDE.md / AGENTS.md        add the pointer from the agent guide
  .claude/skills/fuda-tasks/   the agent guide
```

The tasks folder is flat: no sub-folders, grouping is done with labels. The board reads
`develop`; watching `main` is optional and adds the "in prod" badge.

## Checklist

1. Create `docs/board/tasks/` and add or migrate tasks ([Task format](/guide/task-format)). An
   agent with the [agent guide](/guide/agent-guide) can migrate an old backlog.
2. Optionally add `stages.md`, `labels.md`, `people.md` and `repos.md` ([Board config files](/guide/board-config)).
3. Copy the workflow into `TASKS.md` and add the agent guide and the `CLAUDE.md` pointer.
4. Make CI skip changes that only touch `docs/board/` ([Workflow](/guide/workflow#ci-cd)).
5. Try it locally ([Run locally](/guide/run-locally)) and fix anything listed under Problems.
6. Name the repository `fuda-<something>`, install the fuda GitHub App on it and deploy fuda
   ([Self-host](/guide/self-host)). Everyone who can read the repository sees it as a Board.

A repository with no task files, or with no `develop` branch yet, is not an error. It shows an empty board with the default stages
and the hint "No Tasks yet: add files in `tasks/`".
