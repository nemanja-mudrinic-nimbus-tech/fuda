# Moving existing Task files into a tasks repository

Type: grilling
Status: resolved
Blocked by: 

## Question

Existing Task files live in `docs/board` of a code repository. How are they moved into a new `fuda-` tasks repository and mounted back as a submodule? Keep git history (filter-repo / subtree split) or start fresh? Is it a documented manual step, a script, or a fuda command?

## Answer

Decided with the user, 2026-10-08.

- **Keep the history.** Use `git subtree split --prefix=docs/board` in the code repository, then push that branch as `main` of the new `fuda-` tasks repository. It is built into git, so nothing extra to install. (Rejected: start fresh; `git filter-repo`, because it needs an extra install.)
- **A manual checklist, not a command.** Steps in `CONTRIBUTING.md`: create the `fuda-` repo, subtree split and push, remove `docs/board` from the code repository, add the submodule at `docs/board` with `branch = main`. It runs only 1 to 3 times. (Rejected: a script or fuda command.)
- **Old links:** no redirects. The path `docs/board/...` stays the same through the submodule.
- Covered by the tasks-repository ADR; no separate ADR.
