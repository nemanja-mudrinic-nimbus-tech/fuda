# fuda

A kanban board and doc reader over a board repository's markdown task files (`docs/board/tasks/*.md` on `main`). It
writes two things back, a Move (`status`) and an Assign (`owner`), as commits with the user's own
token. One Go core serves the API and the embedded React app, as a web server and a desktop app.

## Layout

- `api/` — Go module. `cmd/fuda` wires config and server, `cmd/fuda-desktop` the Wails app; packages
  in `internal/` by responsibility: `config`, `board` (registry, service, sync, cache, rules, Move
  and Assign, `Source` interface), `taskfiles` (parses and edits a repo's task and board-config
  files), `source/{azure,github,local}`, `login` (device flow and OAuth), `keychain`, `workspace`
  (hosts and local folders to Boards), `markdown`, `httpapi`, `web` (embedded SPA), `guide`
  (embedded Guide pages).
- `client/` — Vite + React + TypeScript app. Builds into `api/internal/web/dist`; never contains Go.
- Tickets, architecture, decisions, ADRs, glossary: not in this repo. They live in the fuda board
  repository `nimbus-tech-io/fuda-tickets` (tickets in `docs/board/tasks/`, the rest in `docs/`:
  `architecture.md` (how it works, with diagrams), `decisions.md` (why, plus TBD items with
  proposals), `adr/` (one decision each), `glossary.md` (the domain words and where they live in the
  code), `retro/`, `agents/`). Its checkout is the `fuda board:` line in `CLAUDE.local.md`
  (gitignored); no line: ask the human for the path. Read the docs there before structural
  changes. `CONTRIBUTING.md` — setup, rules, PRs.

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
- The Makefile uses only `go`, `pnpm` and `docker`. The macOS targets also use `codesign` and `hdiutil`,
  which ship with macOS and the Xcode Command Line Tools.
- Update docs in the same work as the behaviour they describe: README and Guide pages in the code
  PR here, `<fuda board>/docs/architecture.md` and `decisions.md` as a commit in the fuda board
  repository. `docs/` there states only what is true in the code; decided-but-unbuilt
  items go under TBD in `decisions.md`, with a proposal.

## Personal overrides

Optional and per developer: if `personal-agent.md` exists next to this file, read it and apply it on
top of these rules. It is gitignored.

@personal-agent.md

## Agent skills

### Issue tracker

Tickets live on the fuda board `nimbus-tech-io/fuda-tickets`. Your checkout's path is the `fuda board:` line in `CLAUDE.local.md`; follow `<board>/docs/agents/issue-tracker.md` for every ticket operation. Put the task id in the branch name and the pull request title.

### Triage labels

The five triage roles are statuses with the same names. See the issue tracker doc.

### Domain docs

Single-context: the glossary is `<board>/docs/glossary.md` and the ADRs are in `<board>/docs/adr/`. See `<board>/docs/agents/domain.md`.
