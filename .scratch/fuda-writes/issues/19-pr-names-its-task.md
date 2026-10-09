# How a code PR names its Task

Type: grilling
Status: resolved
Blocked by: 

## Question

Tasks live in their own repository, so a code PR no longer edits a Task file. fuda still shows In review from open code PRs. How does a code PR say which Task it delivers (Task id in title, body, branch, a label)? How does fuda know which code repositories belong to a tasks repository (a config file in the tasks repository)? Does this read use the user's token? Does `merged` still get set, and by whom?

## Answer

Decided with the user, 2026-10-08.

- **A PR names its Task by id.** fuda finds Task ids (like `SS-12`) in the PR title or branch name. Every id found counts: a PR naming two Tasks puts both In review.
- **Code repos are listed in the board config** of the tasks repository (for example `code_repos: [org/app]`). No guessing from names.
- **Any open PR counts**, whatever its base branch.
- **Read with the user's token.** No server token. A user who cannot see a code repository sees no In review from it.
- **`merged` is set by a person.** When the PR closes, the card goes back to its `status` Stage; the person Moves it to Merged. fuda writes nothing on its own.
- Replaces today's "Delivers" rule in `docs/architecture.md` (the PR's version of the Task file). The final proposal must say so.
- No ADR: easy to reverse.
