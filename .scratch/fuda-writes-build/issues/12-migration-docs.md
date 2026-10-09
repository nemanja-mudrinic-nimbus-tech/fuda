# 12: Migration docs

**What to build:** `CONTRIBUTING.md` explains how to move `docs/board` into a `fuda-` tasks repository with its history (`git subtree split --prefix=docs/board`, push as `main`, replace `docs/board` with a submodule with `branch = main`), and how to keep it current with `git submodule update --remote docs/board`. It gives the rule for agents: update the submodule before reading Tasks; commit Task edits inside `docs/board`.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] Checklist and submodule section in `CONTRIBUTING.md`, linked from the README
- [ ] Agent rule text ready to copy into a code repository's `CLAUDE.md`/`AGENTS.md`

Spec: `.scratch/fuda-writes/spec.md`.
