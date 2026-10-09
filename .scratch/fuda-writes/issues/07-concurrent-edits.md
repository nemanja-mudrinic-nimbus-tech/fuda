# Two changes to the same Task at once

Type: grilling
Status: resolved
Blocked by: 02, 05

## Question

Two people Move the same Task within seconds of each other, or a person Moves a Task while a developer pushes an edit to the same file. What does each of them see, which change wins, and how is the loser told? The rule must never lose a developer's commit.

## Answer

Decided with the user, 2026-10-08.

- Every write is a compare-and-swap commit. Fuda never force-pushes, so no developer commit is ever lost.
- **Other lines or files changed first:** fuda reads the newest file, applies the same one-line edit again, and retries. The user is not told.
- **The same field changed first** (`status` for a Move, `owner` for an Assign), by a person or by a developer push: **first wins**. Fuda refuses the write and does not overwrite a value the user never saw.
- **The loser is told:** the card goes back to its real place, with a short message, for example "Ben moved this to Review just now".
- No ADR: easy to reverse.
