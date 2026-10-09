# 03: Move on GitHub

**What to build:** A person drags a card to another Stage. The card moves at once with a saving dot and cannot be dragged again until the save ends. fuda changes only the `status` line (and adds `claimed` set to today the first time the Task leaves the first Stage) and writes one compare-and-swap commit as the user. Other-line changes are retried silently. If `status` changed first, first wins: the card goes back with a message like "Ben moved this to Review just now". On any failure the card goes back with the reason; an expired login shows a "log in again" link. The browser warns when leaving mid-save.

**Blocked by:** 02: GitHub login and the Board picker

**Status:** done

- [x] Pure edit tests: status change, `claimed` added once and never changed, other bytes untouched
- [x] Board service tests with an in-memory fake host: happy path, other-line retry, same-field conflict, never force
- [x] HTTP smoke test covers a Move and a conflict
- [ ] Another user sees the Move within about 10 seconds
- [x] ADR 0001 and 0002 accepted; "Read-only" in `docs/decisions.md` replaced

Spec: `.scratch/fuda-writes/spec.md`.
