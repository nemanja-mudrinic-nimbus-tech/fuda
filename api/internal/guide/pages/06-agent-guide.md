# Agent guide

Where the tasks live decides which block you copy.

**Tasks in a tickets repository.** Copy the skill below into `.claude/skills/fuda-tasks/SKILL.md`
(or an `AGENTS.md` section) of the tickets repository, and add this pointer to each code
repository's `CLAUDE.md`:

```markdown
## Tasks
Work is tracked as task files in `docs/board/tasks/` of the tickets repository (rules: `docs/board/TASKS.md`; columns,
labels and people: `docs/board/stages.md`, `labels.md`, `people.md`). Before starting work, find
or create its task and claim it on `main` of the tickets repository. After the PR that delivers it
merges, set `status: "merged"` and its PR number there. Follow the `fuda-tasks` skill for every task change.
```

**Tasks in the code repository itself.** The tasks are in the repository's own `docs/board/tasks/`.
Copy the skill into the repository's `.claude/skills/fuda-tasks/SKILL.md` and read "the tickets
repository" in it as this repository. A task change is an ordinary commit. When `main` is protected,
it goes in a pull request, and fuda's Move and Assign fail with a hint to change the task in a pull
request.

## The skill

```markdown
---
name: fuda-tasks
description: Create, claim and move task files in docs/board/tasks so the fuda board stays right. Use before starting any task, when the PR that delivers it merges, and when retesting, blocking, abandoning or archiving one.
---

# Task files for the fuda board

Tasks live in docs/board/tasks/, one file per task, named <ID>-<slug>.md. The frontmatter is the
task's state. Read docs/board/TASKS.md for the rules and docs/board/labels.md, stages.md and
people.md for the allowed label values, statuses and people. `owner` holds a person's `id` from
people.md (your own id when you are listed with `kind: agent`), never a display name. Never invent a status or a label
group that is not there; if one is really new, add it to the config file in the same commit.

Before writing code for a task:
1. Pull `main` of the tickets repository and open the task file.
2. If it has an owner, stop and ask the human.
3. Otherwise set owner, status "in progress" and claimed (today), commit only that file as
   "chore(tasks): claim <ID>", show the commit to the human and push to `main` only after they
   approve.

New work found while working:
- Create a task from the template with the next unused id, status "backlog", added today and a
  type label. Show it before pushing.

After the pull request that delivers the task merges:
- In the tickets repository, set status "merged", add the PR number to pr, and write what was checked under "## Evidence".

Retest, blocked, abandon, done, archive:
- testing (with tester), then validated; on failure back to "in progress" with what failed.
- Done: status "done" and done (today) once the task is validated. The file stays where it is.
- Set or clear blocked_by.
- Abandon: status "backlog", clear owner and claimed, explain why under "## Evidence".
- Archive is later cleanup: `git mv` a done or won't-do file to docs/board/archive/ (keeping its
  folder) and change nothing in the file, then fix every link to it in the same commit.

Activity:
- When you finish a step on a task (claim, spec, build, review, test, merge, block), add one line
  at the end of its "## Activity" section; create the section at the end of the body if it is
  missing. One line per step, in this form (the usage part is optional):
  - 2026-10-27 10:40 · <your name> · <step> · <result> · <usage>
  The time is YYYY-MM-DD HH:MM. The result is one word, for example PASS, READY, CHANGES,
  BLOCKED or FAILED, or WORD: VALUE such as VERDICT: CHANGES (the board shows the value).
- Under the line, add detail lines indented by two spaces when the result alone does not say
  enough: findings, what happens next, a failure reason, or the path or URL of the evidence.
  For example:
  - 2026-10-27 11:05 · reviewer · code review · VERDICT: CHANGES · 40k tokens
    - The Move handler does not check the token.
    - The new parser has no test for an empty section.
    - next: build fixes both findings
    - ~/runs/A1/review.md
- Only add lines. Never change, reorder or delete an existing Activity line.

Commits that change task files touch only docs/board/ and use
"chore(tasks): <add|claim|testing|validated|done|blocked|abandon|archive> <ID>".

Keep files valid: an invalid task file disappears from the board into Problems.
```

## Matt Pocock's skills

[Matt Pocock's skills](https://github.com/mattpocock/skills) (`wayfinder`, `to-tickets`, `to-spec`,
`triage`, `implement`) read and write tickets through an issue-tracker doc. With the doc below they
use the fuda board as their tracker, as they use local markdown or GitHub Issues.

### Board setup

The triage roles are statuses, and a won't-do ticket keeps its own column.
Add them to the board's `stages.md`, and the ticket types to `labels.md`:

```markdown
---
stages:
  - name: Backlog
    status: [backlog, needs-triage, needs-info, ready-for-agent, ready-for-human]
  - name: In progress
    status: [in progress]
  - name: In review
    when: pr-open
  - name: Merged
    status: [merged]
  - name: Testing
    status: [testing]
  - name: Validated
    status: [validated]
  - name: Done
    status: [done]
  - name: Won't do
    status: [wontfix]
---
```

```yaml
groups:
  wayfinder: [research, prototype, grilling, task]
```

### The tracker doc

Copy this into the board repository as `docs/agents/issue-tracker.md`, with your id prefix in
place of `<PREFIX>`. It is the one copy: code repositories point to it.

````markdown
# Issue tracker: fuda board

Tickets, specs and wayfinder maps live as task files in the board repository (`fuda-<project>`, or a code repository's `docs/board/`), on
`main`. fuda shows them as a board. The task file format is in fuda's Guide (Task format); the
delivery lifecycle (claim, merged, testing, done, archive) and the commit messages are in the
`fuda-tasks` skill and `docs/board/TASKS.md`. This doc maps the issue-tracker operations onto them.

## The board checkout

`<board>` below is the developer's own checkout of the board repository (`fuda-<project>`, or the
code repository that holds `docs/board/`):

- In a code repository whose tickets are in another repository, `<board>` is the developer's own
  checkout of that repository: its path is the `fuda board:` line in `CLAUDE.local.md`. No such
  line: ask the human for the path and offer to add it. Task files are written only in `<board>`.
- In the board repository itself, `<board>` is the repository root.

Every operation starts with `git -C <board> pull --rebase` and ends with a commit pushed to `main`
(see Commits).

## Conventions

- One file per ticket: `<board>/docs/board/tasks/<epic>/<ID>-<slug>.md`. The folder is the epic
  (the feature or effort slug), exactly one level below `tasks/`.
- **Ids**: one sequence for the whole repository, `<PREFIX>-N`. A new id is the highest N in
  `tasks/` and `archive/` plus one, read after the pull. Ids are never reused and never restart per
  epic.
- **Status**: the `status` field, only words listed in `docs/board/stages.md`. The triage roles
  (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`) are statuses with
  the same names.
- **Labels**: only groups and values listed in `docs/board/labels.md`.
- **Blocking**: `blocked_by: "<PREFIX>-3, <PREFIX>-9"`, ids only.
- **Links** between tickets are relative links to the file (`[Title](<PREFIX>-12-slug.md)`); fuda
  opens them as tasks. A commit that moves a file to `archive/` also fixes every link to it
  (`rg -l '<ID>-' <board>/docs/board`).
- **Comments** go at the bottom under `## Comments`, each starting with the date and the author.

A new ticket starts with:

```yaml
---
id: "<PREFIX>-N"
title: "One line"
status: "needs-triage"
blocked_by: ""
owner: ""
labels: []
added: "YYYY-MM-DD"
claimed: ""
---
```

## When a skill says "publish to the issue tracker"

Create a ticket file as above, `added` today, `status` the triage role the skill names (`needs-triage`
when it names none). For a set of tickets, number blockers first so `blocked_by` names real ids, and
commit the set together.

## When a skill says "fetch the relevant ticket"

After the pull, `find <board>/docs/board -name '<ID>-*.md'`. Given a title or a bare number, search
the `title` fields and ask when more than one matches.

## Closing a ticket

- **Done**: decision tickets as in Resolve below; implementation tickets through the `fuda-tasks`
  lifecycle, which ends in `status: "done"`.
- **Won't do** (wontfix, out of scope, abandoned): `status: "wontfix"`, clear `owner`, the reason
  under `## Answer`. The file stays in `tasks/<epic>/`.
- **Archive**: later cleanup for finished or won't-do tickets: `git mv` the file to
  `archive/<epic>/`, change nothing in it, and fix every link to it in the same commit.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a file in the epic folder; its **child** tickets are the task
files beside it.

- **Map**: `<board>/docs/board/tasks/<epic>/map.md`, opening with a frontmatter block that holds
  `label: wayfinder:map`; fuda shows it as the epic's map. It has no id. The body is the
  Destination / Notes / Decisions so far / Not yet specified / Out of scope sections.
- **Child ticket**: a ticket in the map's folder, `status: "backlog"`, labels `["wayfinder:<type>"]`
  (`research`/`prototype`/`grilling`/`task`), the question under `## Question`.
- **Blocking**: `blocked_by`. A ticket is unblocked when every ticket it lists has `status: "done"`
  or `status: "wontfix"`, wherever the file is.
- **Frontier**: the tickets in `tasks/<epic>/` with `status: "backlog"`, an empty `owner`, and no
  open blocker; lowest id first.
- **Claim**: set `owner` to the driving dev, `status: "in progress"`, `claimed` today; commit and push
  before any work. A rejected push means someone else wrote first: pull, and if the ticket now has an
  owner, take the next frontier ticket.
- **Resolve**: write the answer under `## Answer`, set `status: "done"` and `done` today, and
  append `- [Title](<file>): <gist>` to the map's Decisions so far. The file stays in
  `tasks/<epic>/`. One commit.
- **Out of scope**: close it as won't-do (above) and add its line to the map's Out of scope.

## Commits

Each change is a small commit in `<board>` that touches only `docs/board/`, named as in the
`fuda-tasks` skill (`chore(tasks): <verb> <ID>`, with `resolve` and `wontfix` as extra verbs), pushed
straight to `main`. When a push is rejected, `git pull --rebase`, check that the ids you created are
still free (renumber the file and its references when one is taken), and push again.
````

### In each code repository

Add the pointer to the `## Agent skills` block of `CLAUDE.md`, replacing a local-markdown or GitHub
tracker entry:

```markdown
### Issue tracker

Tickets live on the fuda board `fuda-<project>`. Your checkout's path is the `fuda board:` line in
`CLAUDE.local.md`; follow `<board>/docs/agents/issue-tracker.md` for every ticket operation.

### Triage labels

The five triage roles are statuses with the same names. See the issue tracker doc.
```

Each developer keeps their own path in `CLAUDE.local.md`, which Claude Code loads next to
`CLAUDE.md`; add `CLAUDE.local.md` to `.gitignore`:

```markdown
fuda board: /Users/you/projects/fuda-<project>
```
