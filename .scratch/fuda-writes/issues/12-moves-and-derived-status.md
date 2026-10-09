# How Moves interact with In review from PRs and the archive

Type: grilling
Status: resolved
Blocked by: 

## Question

fuda today derives "In review" from open PRs and archives `done` Tasks. When a user can Move a card: may they Move a card into or out of a PR-derived Stage? What does that write? Can a user Move an archived `done` Task back? Read `docs/architecture.md` for the current rules.

## Answer

Decided with the user, 2026-10-08.

- Found: with Tasks in their own repository, code PRs no longer change Task files. Today's match (a PR's version of the Task file) stops working.
- **In review stays PR-derived.** fuda also reads open PRs in the code repository. How a PR names its Task is the ticket "How a code PR names its Task".
- **Locked while a PR is open.** A card in In review because of an open PR cannot be dragged. No card can be dropped into a PR-derived Stage by hand. When the PR closes, the card goes back to the Stage its `status` says.
- **Archived Tasks are read-only in the app.** No Move into or out of the archive folder. Archiving stays a file move in git. Every app write stays a one-line edit.
- No ADR: easy to reverse.
