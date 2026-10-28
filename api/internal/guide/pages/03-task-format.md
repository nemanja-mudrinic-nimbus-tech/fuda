# Task format

One file per task in `docs/board/tasks/`. The filename starts with the id: `<ID>-<slug>.md` or
`<ID>.md`.

```markdown
---
id: "SS-12"
title: "Invitation email failure rolls back the insert"
status: "backlog"
blocked_by: ""
owner: ""
tester: ""
labels: ["type:bug", "area:api", "epic:invitations"]
added: "2026-10-01"
claimed: ""
pr: []
done: ""
---
# SS-12: Invitation email failure rolls back the insert

What, why, where (files and symbols).

## Done when
The check that proves it is finished.

## Evidence
What was checked or changed, with dates and commit shas.
```

## Fields

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Unique, never reused |
| `title` | yes | One line |
| `status` | yes | Decides the column. Default set: `backlog`, `in progress`, `merged`, `testing`, `validated`; `done` only in `archive/` |
| `blocked_by` | no | Task ids (`"SS-3, SS-9"`) or text. Non-empty means blocked, at any status |
| `owner`, `tester` | no | One or more names (`"Ron, Chinmay"` or a list) |
| `labels` | no | A list of `group:value` strings, e.g. `type:bug`, `epic:billing` |
| `added`, `claimed`, `done` | no | `YYYY-MM-DD`; `added` is what Insights and the Added filter use |
| `pr` | no | Pull request numbers, or a branch name |
| anything else | no | Custom fields, shown in the task panel, not used for columns or filters |

Values are any valid YAML. Quoted JSON style, as above, is the safest for simple checkers.

## Body

The body is GitHub-flavoured markdown, including tables and mermaid diagrams. These sections
are recommended:

- `## Done when`: shown first in the task panel.
- `## Evidence`: what was checked, with dates and commit shas.
- `## Group context`: once per label group, for background that many tasks share.
- `## Answer`: on a question once it is answered.
- `## Activity`: one line per event, added at the end. The task panel shows it as a timeline,
  newest on top, instead of a list in the body.

## Activity

Each line of `## Activity` is one entry:

```markdown
## Activity

- 2026-10-27 09:12 · factory · spec · READY · 12k tokens
- 2026-10-27 10:40 · builder · build · PASS
- 2026-10-27 11:05 · reviewer · code review · VERDICT: CHANGES
  - The Move handler does not check the token.
  - next: build fixes the finding
  - ~/runs/A1/review.md
```

The parts are `<YYYY-MM-DD HH:MM> · <actor> · <event> · <result> · <usage>`; the usage may be
missing. The panel shows each entry as a row: the actor with an avatar, the event, the result as a
badge, the time relative to now ("just now", "12 min ago", "3 h ago", "yesterday", then the date)
with the full time on hover, and the usage at the end. The badge shows only the result word: for
`WORD: VALUE` it shows the value (`VERDICT: PASS` shows `PASS`), otherwise the whole result. The
badge is green for `PASS`, `READY`, `COMPLETE` and `auto_merge`, amber for `CHANGES`, `GAPS` and
`needs_human`, red for `BLOCKED`, `FAILED` and `HIT`, and grey for any other result. A line in
another form shows as an entry with its text only; it is never a Problem.

Lines indented under an entry (two spaces, then `- `) are its detail lines: findings, what happens
next, evidence or a failure reason. They are never entries of their own. The panel shows them under
the row, the first 3 and a "show N more" control for the rest. A detail line that starts with `~/`
or `/` shows with a copy button; a URL shows as a link.

A column with `activity: true` in `stages.md` shows the last entry on its cards as
`<actor> · <event>` with the result badge, without detail lines.

Task ids written anywhere in a body become links, and the panel lists which tasks block or
mention each other.

## When a file is wrong

fuda skips the file and lists it under **Problems** with the reason: broken YAML, a missing
`id`, `title` or `status`, a filename that does not start with the id, `done` outside the archive,
or a duplicate id. Everything else on the board keeps working.
