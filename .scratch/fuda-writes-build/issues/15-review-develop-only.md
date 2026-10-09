# 15: Review whether everything stays on the develop branch

**What to build:** Nothing yet. This is a review that ends in a decision. fuda reads and writes Tasks on one branch, `develop` (`WorkBranch`), and reads `main` only for the "in prod" badge. Decide whether that is still right now that Moves and Assigns write commits.

Questions to answer:

- Should Moves and Assigns always commit to `develop`, or to the repository's default branch, or to a branch the Board config names?
- Does a repository with only `main` work well? Ticket 08 already shows an empty Board when `develop` is missing.
- Is the "in prod" badge from `main` still worth a second read, now that Tasks live in their own repository?
- Does protected-branch policy on `develop` break the write (a direct commit is refused)? What should the user see?

Write the answer in `docs/decisions.md`. If the answer changes behaviour, split it into new tickets.

**Blocked by:** 05: Assign on GitHub

**Status:** needs-triage

- [ ] The four questions above are answered
- [ ] The decision is recorded in `docs/decisions.md`
- [ ] Follow-up tickets are made if behaviour must change

Spec: `.scratch/fuda-writes/spec.md`.
