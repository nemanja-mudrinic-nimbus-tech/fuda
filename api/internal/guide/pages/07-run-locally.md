# Run locally

Point fuda at a checkout on disk to see the board before deploying, or to preview task edits
before you push them.

```sh
git clone <fuda> && cd fuda
cp .env.example .env          # set FUDA_LOCAL_PATH to your repository
make dev                      # API on :8080, app on http://localhost:5173
```

With the local source the board shows your working tree, uncommitted edits included. Move and
Assign change the task file on disk; you commit them. The folder can have `docs/board/tasks/` or
`tasks/`. In the desktop app, use File → Open folder…. If the
checkout has `origin/main` and `FUDA_WATCH_MAIN=true`, "in prod" works too.

| Setting | Use |
|---|---|
| `FUDA_SOURCE=local`, `FUDA_LOCAL_PATH=…` | Read a checkout; no login |
| `FUDA_SOURCE=github` and the GitHub App settings | Log in with GitHub, as in production |

The browser polls the Board about every 5 seconds, so edits to the working tree show up on their
own. The [Self-host guide](/guide/self-host) shows how to create the GitHub App.

Prerequisites: Go 1.27+, Node 24+ and pnpm 10. `make check` runs every lint and test.
