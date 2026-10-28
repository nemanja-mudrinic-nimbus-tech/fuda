# Agent guide

Copy the block below into `.claude/skills/fuda-tasks/SKILL.md` (or an `AGENTS.md` section), and
add this pointer to your `CLAUDE.md`:

```markdown
## Tasks
Work is tracked as task files in `docs/board/tasks/` (rules: `docs/board/TASKS.md`; columns,
labels and people: `docs/board/stages.md`, `labels.md`, `people.md`). Before starting work, find
or create its task and claim it on develop. The PR that delivers it sets `status: "merged"` and
its PR number. Follow the `fuda-tasks` skill for every task change.
```

## The skill

```markdown
---
name: fuda-tasks
description: Create, claim and move task files in docs/board/tasks so the fuda board stays right. Use before starting any task, when opening the PR that delivers it, and when retesting, blocking, abandoning or archiving one.
---

# Task files for the fuda board

Tasks live in docs/board/tasks/, one file per task, named <ID>-<slug>.md. The frontmatter is the
task's state. Read docs/board/TASKS.md for the rules and docs/board/labels.md, stages.md and
people.md for the allowed label values, statuses and names. Never invent a status or a label
group that is not there; if one is really new, add it to the config file in the same commit.

Before writing code for a task:
1. Pull develop and open the task file.
2. If it has an owner, stop and ask the human.
3. Otherwise set owner, status "in progress" and claimed (today), commit only that file as
   "chore(tasks): claim <ID>", show the commit to the human and push to develop only after they
   approve.

New work found while working:
- Create a task from the template with the next unused id, status "backlog", added today and a
  type label. Show it before pushing.

In the pull request that delivers the task:
- Set status "merged", add the PR number to pr, and write what was checked under "## Evidence".

Retest, blocked, abandon, archive:
- testing (with tester), then validated; on failure back to "in progress" with what failed.
- Set or clear blocked_by.
- Abandon: status "backlog", clear owner and claimed, explain why under "## Evidence".
- Never write status "done" except when moving a file to docs/board/archive/.

Activity:
- When you finish a step on a task (claim, spec, build, review, test, merge, block), add one line
  at the end of its "## Activity" section; create the section at the end of the body if it is
  missing. One line per step, in this form (the usage part is optional):
  - 2026-10-27 10:40 · <your name> · <step> · <result> · <usage>
  The time is YYYY-MM-DD HH:MM. The result is one word, for example PASS, READY, CHANGES,
  BLOCKED or FAILED.
- Only add lines. Never change, reorder or delete an existing Activity line.

Commits that change task files touch only docs/board/ and use
"chore(tasks): <add|claim|testing|validated|blocked|abandon|archive> <ID>".

Keep files valid: an invalid task file disappears from the board into Problems.
```
