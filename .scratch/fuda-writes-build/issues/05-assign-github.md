# 05: Assign on GitHub

**What to build:** A person picks the Owners of a Task from the short names in `people.md` (or names already on the board when there is no `people.md`). fuda writes `owner` in the file's existing style (comma text or YAML list); a new `owner` is comma text after `status`; removing every Owner deletes the line. Same compare-and-swap and conflict rule as Move, on the `owner` field.

**Blocked by:** 03: Move on GitHub

**Status:** resolved

- [x] Pure edit tests for comma style, YAML list style, new line, delete line
- [x] Board service tests: Assign happy path and same-field conflict
- [x] Owner picker in the UI with saving and rollback like Move
- [x] Guide page updated

Spec: `.scratch/fuda-writes/spec.md`.
