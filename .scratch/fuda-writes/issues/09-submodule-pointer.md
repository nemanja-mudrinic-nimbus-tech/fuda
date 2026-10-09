# How developers and agents keep the tasks submodule up to date

Type: grilling
Status: resolved
Blocked by:

## Question

The code repository's submodule pointer to the tasks repository gets old after every Move. How do developers and agents always see the newest Tasks? Options to weigh: `branch =` in `.gitmodules` plus `git submodule update --remote`, `submodule.recurse`, a git hook, an agent rule in `CLAUDE.md`/`AGENTS.md`, or never bumping the pointer at all. Where is this explained for people (README, `CONTRIBUTING.md`, a Guide page) and for agents? How does a developer commit a Task edit from inside the submodule?

## Answer

Decided with the user, 2026-10-08.

- **Track the branch, never bump the pointer.** `.gitmodules` sets `branch = main` for the tasks submodule. People and agents run `git submodule update --remote docs/board` to see the newest Tasks. Nobody commits the pointer in the code repository; an old pointer is expected and harmless. (Rejected: a hook or bot that bumps the pointer: noise commits and merge conflicts.)
- **Fuda is the main way to Move and Assign.** Manual edits stay supported: `cd docs/board`, `git checkout main && git pull`, edit, commit, push. Agents use the same flow. Fuda's compare-and-swap writes already handle a push that lands between its read and write.
- **Where it is explained.** People: a short section in `CONTRIBUTING.md`, linked from the README. Agents: a rule in the code repository's `CLAUDE.md`/`AGENTS.md` (update the submodule before reading Tasks; commit Task edits inside `docs/board`). Fuda's Guide pages: no change, fuda does not use the submodule.
- **Task-creation skill:** an agent skill for other people's agents to create one or more Tasks (instead of pasting markdown in chat). Ruled out of scope for this map; a separate effort. Leaning: it writes through git (clone, commit, push).
- Covered by the tasks-repository ADR; no separate ADR.
