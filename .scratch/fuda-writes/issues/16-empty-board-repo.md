# A fuda- repository with no Task files or no board config

Type: grilling
Status: resolved
Blocked by: 10

## Question

A repo named `fuda-*` is a Board. What if it has no Task files, or no board config? Show an empty board with defaults, or a Problem? Is this how a new Board is started?

## Answer

Decided with the user, 2026-10-08 (grilling).

- **No Task files** (empty repo, or no `tasks/` folder): an empty board with default Stages and a one-line hint, "No Tasks yet: add files in `tasks/`". Not a Problem.
- **Starting a new Board:** create an empty `fuda-` repo on the host. It shows up as an empty Board at once. The first Task is added in git. No "New Board" button in fuda. A short step in `CONTRIBUTING.md`. (Rejected: a button; it needs repo-create permission and the app does not create Tasks.)
- **Invalid board config:** keep today's rule. The file is a Problem and is ignored; the board uses defaults. (Briefly chose "block the board"; reversed because it breaks "repo content never breaks fuda".)
- No ADR: small and easy to reverse.
