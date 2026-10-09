# 07: In review from code PRs

**What to build:** The board config lists code repositories (`code_repos`). fuda lists open PRs in each with the user's token, on any base branch, and finds Task ids in the PR title and branch name. Every Task named is shown in the PR-derived Stage and is locked: it cannot be dragged, and nothing can be dropped into that Stage by hand. When the PR closes, the card returns to its `status` Stage. fuda never sets `merged`. The old match by the PR's version of the Task file is removed.

**Blocked by:** 03: Move on GitHub

**Status:** done

- [x] Board service tests: id in title, id in branch, two ids, any base, PR closed returns card
- [x] Locked card refuses Move (test) and does not drag (UI)
- [x] A user who cannot see a code repo still sees the Board, without its In review
- [x] `docs/decisions.md` In review rule and `docs/architecture.md` updated

Spec: `.scratch/fuda-writes/spec.md`.
