# fuda

A read-only kanban board and doc reader over a git repo's markdown task files
(`docs/board/tasks/*.md`). One Go binary serves the API and the embedded React app.

## Layout

- `api/` — Go module. `cmd/fuda` wires config and server; packages in `internal/` by responsibility:
  `config`, `board` (service, sync, cache, rules, `Source` interface), `taskfiles` (parses a repo's
  task and board-config files), `source/{azure,github,local}`, `markdown`, `httpapi`, `web` (embedded
  SPA), `guide` (embedded Guide pages).
- `client/` — Vite + React + TypeScript app. Builds into `api/internal/web/dist`; never contains Go.
- `docs/` — `architecture.md` (how it works, with diagrams) and `decisions.md` (why, plus TBD items
  with proposals). Read them before structural changes. `CONTRIBUTING.md` — setup, rules, PRs.

## Commands

- `make dev` — API on :8080 and Vite on :5173 (proxies `/api`). Needs `.env` (see `.env.example`).
- `make check` — Go lint + tests, client lint + types + tests. Run before every commit.
- `make build` — client then binary at `bin/fuda`. `make docker` — the image.

## Rules

- Self-explanatory names. No comments that restate code — no doc comments on functions or methods,
  in Go or TypeScript. A comment only for a non-obvious why.
- Handlers are thin; behaviour lives in `board`. Interfaces are declared where they are used.
- No speculative abstractions. Unit tests where the logic is; small fixtures in `testdata/`.
- Repo content never breaks fuda: invalid files become Problems, a failed sync keeps the last snapshot.
- The Makefile uses only `go`, `pnpm` and `docker`.
- Update docs (README, Guide pages, `docs/architecture.md`, `docs/decisions.md`) in the same change
  as the behaviour they describe. `docs/` states only what is true in the code; decided-but-unbuilt
  items go under TBD in `decisions.md`, with a proposal.

## Personal overrides

Optional and per developer: if `personal-agent.md` exists next to this file, read it and apply it on
top of these rules. It is gitignored.

@personal-agent.md

## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

The five default labels: needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
