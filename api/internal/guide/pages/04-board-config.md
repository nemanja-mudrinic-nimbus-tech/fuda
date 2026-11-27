# Board config files

Four optional files in `docs/board/`. Each is normal markdown with a YAML frontmatter block at
the top that fuda reads; the rest of the file is for people. A missing file means fuda derives
the same thing from the tasks. An invalid block is listed under Problems and ignored.

fuda reads Board format 0.2. A Board is the layout of these files and of the task files. The Board
format changes only when the layout does, not with every fuda release; fuda keeps reading older Boards.

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
    status: [in review]
    when: pr-open
  - name: Merged
    status: [merged]
  - name: Testing
    status: [testing, in qa]
  - name: Validated
    status: [validated]
  - name: Done
    status: [done]
---
```

- The list order is the column order. A column can hold several statuses; the card then shows its
  own status icon.
- `when: pr-open` is the In review column: a task that an open pull request names shows there,
  whatever its status, and is locked. With a `status` list as well, a card can also be dragged
  there, which writes the first status. Without one, only open pull requests fill it.
- `activity: true` makes the column's cards show the latest entry of the task's `## Activity`
  section in short form: `<actor> · <event>`. A card without that section looks the same as in
  any other column.
- `when: done` adds a Done column with the archived tasks whose `done` date is within the last
  `days` days, today included, newest first whatever the board's sort. Without `days` the window is 14 days. An archived
  task without a valid `done` date (`YYYY-MM-DD`) is not in the column. Older tasks show only on
  the Archive tab. The column is always the last one and takes no status list. Its cards cannot
  be moved or assigned, and nothing can be dropped on it. The filters apply to it as to the other
  columns. Without a `when: done` entry, archived tasks show only on the Archive tab.
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
  a task but not listed still shows on the card, it just has no filter entry. Epic is the
  exception: every epic folder with a task is in the Epic filter, listed in `groups` or not.
- `colors` is optional. Without it, fuda gives every value in a group its own colour from a
  palette, and `type` keeps fixed defaults (bug red, feat blue, …).
- `theme:` and `epic:` are treated as the same group, shown as Epic. Both labels are deprecated:
  put the task in `tasks/<epic>/` instead, which sets `epic:<folder>`.

## people.md: people and agents

```markdown
---
people:
  - id: nemanja
    name: Nemanja Mudrinic
    aliases: [Nemanja]
    github: nmudrinic
  - id: research-agent
    name: Research agent
    kind: agent
---
```

| Key | Required | Meaning |
| --- | --- | --- |
| `id` | yes | The stable value tasks store in `owner` and `tester`, such as `owner: nemanja`. Unique. |
| `name` | yes | The name fuda shows. |
| `aliases` | no | Extra words the people search finds. Not used to match `owner`. |
| `github` | no | The GitHub login. Shows the person's avatar, maps the logged-in user to them ("Assign to me", "My tasks"), and shows them as the author of an open pull request. |
| `kind` | no | `human` (default) or `agent`. Agents get a bot badge; the Owner filter can pick Humans or Agents. |

With the file, `owner` and `tester` hold ids, matched exactly. A value that is not an id is listed
under Problems for that task, and the card shows it as written. A missing or repeated `id`, or an
unknown `kind`, makes the whole file a Problem; fuda then derives people from the tasks. Assign
writes ids, and only ids listed here.

Without the file, people come from the names used in tasks, matched case-insensitively, and a
unique first name is matched to the full name.
The Owner filter also offers "No owner" for the cards that have no owner.

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
