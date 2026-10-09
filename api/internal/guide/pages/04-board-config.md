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
  - name: In review
    when: pr-open
  - name: Merged
    status: [merged]
  - name: Testing
    status: [testing, in qa]
  - name: Validated
    status: [validated]
---
```

- The list order is the column order. A column can hold several statuses; the card then shows its
  own status icon.
- `when: pr-open` is the derived column: a task that an open pull request names shows there,
  whatever its status.
- A status no column lists gets its own column at the end, marked unknown.

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
matched to the full name.

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
