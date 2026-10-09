# fuda

A kanban board and doc reader for teams that keep their tasks as markdown files in git.

Each task is one file in `docs/board/tasks/` with YAML frontmatter (`status`, `owner`, `labels`, …).
fuda reads them from GitHub, Azure DevOps (log in with your Microsoft account) or a local checkout, and shows them as a board with filters,
a task reader and the repo's docs. People (and their AI agents) move tasks with ordinary commits.
On GitHub a person can also drag a card to another Stage: fuda commits that one `status` change as
the logged-in user, and pick a Task's Owners from the short names in `people.md`.

> Status: early development, in use on two repositories.

## Documentation

- **Using fuda on your repository:** the Guide inside the app, or
  [`api/internal/guide/pages`](api/internal/guide/pages): set up a repo, task format, board
  config, workflow, the agent guide, self-hosting.
- **How it works:** [docs/architecture.md](docs/architecture.md).
- **Why it works that way:** [docs/decisions.md](docs/decisions.md).
- **Changing fuda:** [CONTRIBUTING.md](CONTRIBUTING.md).
- **Moving `docs/board` into a `fuda-` repository:** [CONTRIBUTING.md](CONTRIBUTING.md#move-docsboard-into-a-fuda--repository).

## Quick start

```sh
cp .env.example .env      # point FUDA_LOCAL_PATH at any checkout that has docs/board/tasks
make dev                  # API on :8080 and the app on http://localhost:5173
```

## Run the image

```sh
docker run -p 8080:8080 -v fuda-data:/data \
  -e FUDA_GITHUB_CLIENT_ID=<GitHub App client id> fuda
```

Keep the volume: fuda saves its login cookie key there, and a new key logs everyone out. Serve fuda
over HTTPS; login cookies are `Secure` except on localhost.

People log in with the fuda GitHub App. fuda lists every `fuda-` repository they can read and the
App is installed on, and reads each Board with that person's own token. There is no server token and
no shared password. Boards open at `/github/<owner>/<repo>/`; `/` lists them, or opens the only one.
Set up the App in the [Self-host guide](api/internal/guide/pages/08-self-host.md).

## Desktop app

Download the archive for your system from the latest GitHub Release, unpack it and open `fuda.app`
(macOS) or `fuda-desktop.exe` (Windows). Log in with GitHub: fuda shows a code and opens GitHub, where
you type it. Azure DevOps works the same way, and you can be logged in to both. Each login token stays in your
OS keychain, and you can log out of one host or of all from the account menu in the top bar. The app updates itself from new releases.

The builds are not signed, so your system warns the first time:

- **macOS:** "fuda cannot be opened". Right-click `fuda.app`, choose Open, then Open again. Or
  run `xattr -dr com.apple.quarantine fuda.app`.
- **Windows:** SmartScreen says "Windows protected your PC". Choose More info, then Run anyway.

## License

MIT
