# What a Move and an Assign change in a Task file

Type: grilling
Status: resolved
Blocked by:

## Question

When a person Moves a Task, which frontmatter fields change (`status`, and `claimed` or other dates)? When they Assign, how is `owner` written (one name, a list, the person's name from `people.md`)? Does order inside a Stage matter and need storing? The edit must keep the rest of the file byte-for-byte, so developers see a one-line diff.

## Answer

Decided with the user, 2026-10-08.

- **Move** changes `status`. The first time a Task leaves the first Stage, fuda also sets `claimed` to today. If `claimed` already exists, fuda never changes or deletes it, also when the Task moves back.
- **Assign** writes `owner` in the style the file already uses: a comma text (`"Ron, Chinmay"`) or a YAML list. A new `owner` is a comma text. Removing every Owner deletes the `owner` line.
- **Names** are the short names (aliases) from `people.md`. With no `people.md`, the names already seen on the board.
- **Order inside a Stage** is not stored. A Move only changes the Stage.
- **Edits are line-level.** Only the changed line changes. A new field (`claimed`, `owner`) goes on a new line right after `status`. The rest of the file stays byte-for-byte.
- No ADR: easy to reverse.
