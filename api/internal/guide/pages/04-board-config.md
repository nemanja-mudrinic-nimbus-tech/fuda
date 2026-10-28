# Board config files

Four optional files in `docs/board/`. Each is normal markdown with a YAML frontmatter block at
the top that fuda reads; the rest of the file is for people. A missing file means fuda derives
the same thing from the tasks. An invalid block is listed under Problems and ignored.

## stages.md: the columns

```markdown
---
stages:
  - name: Backlog
    status: [backlog]
  - name: In progress
    status: [in progress]
    activity: true
  - name: In review
    when: pr-open
  - name: Merged
    status: [merged]
  - name: Testing
    status: [testing, in qa]
  - name: Validated
    status: [validated]
  - name: Done
    when: done
    days: 14
---
```

- The list order is the column order. A column can hold several statuses; the card then shows its
  own status icon.
- `when: pr-open` is the derived column: a task that an open pull request names shows there,
  whatever its status.
- `activity: true` makes the column's cards show the latest entry of the task's `## Activity`
  section in short form: `<actor> · <event>`. A card without that section looks the same as in
  any other column.
- `when: done` adds a Done column with the archived tasks whose `done` date is within the last
  `days` days, today included, newest first whatever the board's sort. Without `days` the window is 14 days. An archived
  task without a valid `done` date (`YYYY-MM-DD`) is not in the column. Older tasks show only on
  the Archive tab. The column is always the last one and takes no status list. Its cards cannot
  be moved or assigned, and nothing can be dropped on it. The filters apply to it as to the other
  columns. Without a `when: done` entry, done tasks show only on the Archive tab.
- `days` must be a positive whole number. Any other value is listed under Problems and the column
  uses 14. Only one `when: done` entry counts; a second one is listed under Problems and ignored.
- A status no column lists gets its own column, marked unknown, before the Done column.

## labels.md: label groups, values and colours

```markdown
---
groups:
  type:  [bug, feat, impr, refactor, test, chore, docs, question]
  epic:  [billing, onboarding]
  area:  [api, web, db, infra]
colors:
  "type:bug": "#ef4444"
  "epic:billing": "#22c55e"
---
```

- Each group becomes a filter that offers **only these values**, in this order. A value used in
  a task but not listed still shows on the card, it just has no filter entry.
- `colors` is optional. Without it, fuda gives every value in a group its own colour from a
  palette, and `type` keeps fixed defaults (bug red, feat blue, …).
- `theme:` and `epic:` are treated as the same group, shown as Epic.

## people.md: people and aliases

```markdown
---
people:
  - name: Nemanja Mudrinic
    aliases: [Nemanja]
---
```

The owner and tester filters list only these names, and aliases fold "Nemanja" into "Nemanja
Mudrinic". Without the file, people come from the names used in tasks, and a unique first name is
matched to the full name. The owner filter also offers "No owner" for the cards that have no owner.

## repos.md: code repositories for In review

```markdown
---
code_repos: [acme/app, acme/web]
---
```

fuda lists the open pull requests of each repository, on any base branch, with the logged-in
person's token. A task whose id is in a pull request's title or branch name (`SS-12` in
`SS-12: add login` or `feature/ss-12-login`) shows in the In review column and is locked. One pull
request can name several tasks. If the person cannot see a repository, the board still loads
without that repository's In review.
