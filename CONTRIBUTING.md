# Contributing

Thanks for helping. fuda is small on purpose; before larger changes read `docs/architecture.md` (how
it works), `docs/decisions.md` (why) and `docs/glossary.md` (the words both use). They live in the
fuda board repository, `nimbus-tech-io/fuda-tickets` (private; ask a maintainer for access), next to
fuda's own tickets, ADRs and retros. Its path on your machine is the `fuda board:` line in
`CLAUDE.local.md` (gitignored).

## Set up

Prerequisites: Go 1.27+, Node 24+, pnpm 10 (`corepack enable`), golangci-lint v2, Docker (for images).

```sh
cp .env.example .env      # point FUDA_LOCAL_PATH at any checkout that has docs/board/tasks
make dev                  # API on :8080 and the app on http://localhost:5173
```

`make dev` runs two processes: the Go API (`make dev-api`) and Vite (`make dev-client`, proxying
`/api` to :8080). Run them in separate terminals if you prefer.

| Command | Does |
|---|---|
| `make check` | Go lint + tests, client lint + format check + types + tests. Run before committing. |
| `make fmt` | Formats Go and client code. |
| `make build` | Builds the client into `api/internal/web/dist`, then the binary `bin/fuda`. |
| `make docker` | Builds the image (`fuda`). |
| `make desktop-macos` | The desktop app for macOS as the ad-hoc-signed bundle `bin/fuda.app`. Needs a Mac with Xcode command line tools. `ARCH=amd64` for Intel. Needs the GitHub App values (see "GitHub App locally"). |
| `make desktop-macos-dmg` | `make desktop-macos`, then the disk image `bin/fuda-desktop_darwin_<arch>.dmg` with `fuda.app` and an Applications link. This is what the release uploads. |
| `make desktop-windows` | The desktop app for Windows in `bin/fuda-desktop.exe`. Cross-builds from any host. |
| `make clean` | Removes build output and the local cache. |

## Hosts

| `FUDA_HOSTS` | Needs |
|---|---|
| `github` (default) | A GitHub App: `FUDA_GITHUB_CLIENT_ID` and `FUDA_GITHUB_APP_SLUG`. People log in with device flow; there is no server token. |
| `local` | `FUDA_LOCAL_PATH`: a checkout on disk. Shows the working tree, uncommitted edits included. No login. A server shows it read-only; only the desktop app writes it (Moves and Assigns change the file on disk, you commit). A folder with `tasks/` also works. On desktop, File → Open folder… does this. |
| `azure` | A Microsoft Entra ID app: `FUDA_AZURE_CLIENT_ID`, `FUDA_AZURE_CLIENT_SECRET`, optionally `FUDA_AZURE_TENANT` (default `organizations`), plus `FUDA_BASE_URL` and `FUDA_COOKIE_SECRET`. People log in; there is no server token. `FUDA_COOKIE_SECRET` must be at least 32 characters. |

`FUDA_HOSTS=github,azure` serves both hosts at once, each with its own login and settings. `local`
stands alone. `FUDA_SOURCE` is the old name of `FUDA_HOSTS` and still works.

All settings are in [`.env.example`](.env.example).

## Entra ID app locally

Register an app in Microsoft Entra ID (App registrations) with:

- A Web redirect URI `http://localhost:5173/auth/azure/callback`, and a client secret.
- API permission: Azure DevOps, delegated `user_impersonation`. Whether people can consent to it
  themselves or an admin must grant it depends on the tenant's consent policy.
- For the desktop app: "Allow public client flows" on, and run it with `FUDA_AZURE_CLIENT_ID`. It
  logs in with the device code and keeps the token in the OS keychain. With `FUDA_GITHUB_CLIENT_ID`
  set too, the app serves both hosts.

## GitHub App locally

Create a GitHub App (Settings → Developer settings → GitHub Apps) with:

- "Enable Device Flow" on, "Expire user authorization tokens" off, and "Request user authorization
  (OAuth) during installation" off. No callback URL and no client secret are needed. Vite proxies
  `/auth` to the API, so the login cookie lands on the page you browse.
- Repository permissions: Contents read, Pull requests read, Metadata read.
- No webhook.

Install it on your `fuda-` repositories, then set `FUDA_GITHUB_CLIENT_ID` and `FUDA_GITHUB_APP_SLUG` (the last part of `github.com/apps/<slug>`) in your local env file. The desktop build embeds both: `make desktop-macos` reads them from the env file, or pass `GITHUB_CLIENT_ID=` and `GITHUB_APP_SLUG=`. The build fails when either is missing.

The browser polls the Board about every 5 seconds. The server asks GitHub for the head commit first
(a conditional request, free when nothing changed) and reads files only when it moved. With the local
source the same poll picks up edits to your working tree.

## Start a new Board

Where the Board lives is your choice. Either create a repository named `fuda-<something>`, or take a code
repository and list it in `FUDA_GITHUB_BOARDS` (`FUDA_AZURE_BOARDS` on Azure DevOps), with its tasks in `docs/board/tasks/`. Install the fuda GitHub App on it.
Push a first commit to its `main` branch: a repository without
`main` shows an error page. Open it in fuda: with no Task files, it shows an empty Board with the default Stages and the hint "No Tasks yet: add
files in `tasks/`". Add Task files under `docs/board/tasks/` and they appear on the next sync. When `main` of a code repository is protected, a Move or Assign fails with a hint to change the task in a pull request.

## Move `docs/board` into its own `fuda-` repository

Optional. A code repository that keeps its Board in `docs/board/` can move it into a `fuda-` repository
of its own and keep the history, so the code repository keeps no copy of the Tasks:

1. Create an empty repository named `fuda-<something>` and install the fuda GitHub App on it.
2. In the code repository, split the history of `docs/board` into a branch:
   ```sh
   git subtree split --prefix=docs/board -b board-split
   ```
3. Push that branch as `main` of the new repository:
   ```sh
   git push git@github.com:<owner>/fuda-<something>.git board-split:main
   ```
4. Remove the folder from the code repository:
   ```sh
   git rm -r docs/board
   git commit -m "Move docs/board to fuda-<something>"
   ```
5. Delete the `board-split` branch and list the code repository in the new Board's `repos.md`.
6. Add the agent rule below to the code repository's `CLAUDE.md` or `AGENTS.md`.

### Agent rule

Copy this into the code repository's `CLAUDE.md` or `AGENTS.md`:

```md
## Tasks

Tasks live in the board repository `fuda-<something>`. Your checkout's path is the `fuda board:` line
in `CLAUDE.local.md` (gitignored). Change tasks only there, following `<board>/docs/agents/issue-tracker.md`.
Put the task id in the branch name and the pull request title.
```

## Desktop app

Set `FUDA_GITHUB_CLIENT_ID` (a GitHub App with "Enable Device Flow" on) and run `bin/fuda-desktop`, or
build it in with `make desktop-macos`. The build reads `FUDA_GITHUB_CLIENT_ID`, `FUDA_GITHUB_APP_SLUG` and `FUDA_AZURE_CLIENT_ID` from your local env file (the one `make dev` loads). `GITHUB_CLIENT_ID=<id>`, `GITHUB_APP_SLUG=<slug>` or `AZURE_CLIENT_ID=<id>` on the `make` line wins. The GitHub id and slug are required: the build fails without them. The Azure id is optional. The build also reads the optional board lists `FUDA_GITHUB_BOARDS` and `FUDA_AZURE_BOARDS` from the env file (`GITHUB_BOARDS=` and `AZURE_BOARDS=` on the `make` line win); a value in the environment at run time wins over the built-in one. The App needs "Expire user
authorization tokens" off: there is no client secret, so fuda cannot refresh a token.

To release, bump `VERSION` in a pull request, then push the matching tag: see [How to release](HOW-TO-RELEASE.md).

### Desktop icon

The master art is `client/public/favicon.svg`. After you change it, remake the files in
`api/cmd/fuda-desktop/` and commit them. The build needs no extra tools. Run from the repository root
on a Mac with Python's Pillow (`pip install pillow`):

```bash
cd api/cmd/fuda-desktop
sips -s format png -z 1024 1024 ../../../client/public/favicon.svg --out icon.png
sips -s format png -z 824 824 ../../../client/public/favicon.svg --out /tmp/art.png
python3 -c "from PIL import Image; a=Image.open('/tmp/art.png').convert('RGBA'); c=Image.new('RGBA',(1024,1024),(0,0,0,0)); c.paste(a,(100,100),a); c.save('icon-macos.png')"
go run github.com/tc-hib/go-winres@v0.3.3 simply --icon icon.png --arch amd64,arm64 --out rsrc
d=$(mktemp -d)/fuda.iconset && mkdir $d
for s in 16 32 128 256 512; do
  sips -z $s $s icon-macos.png --out $d/icon_${s}x${s}.png
  sips -z $((s*2)) $((s*2)) icon-macos.png --out $d/icon_${s}x${s}@2x.png
done
iconutil -c icns $d -o icon.icns
```

`icon.png` is for the window and Windows. `icon-macos.png` has a transparent margin, for the Dock.
`icon.icns` goes in `fuda.app`; `Info.plist` next to it describes the bundle.
The `rsrc_windows_*.syso` files put the icon in `fuda-desktop.exe`; Go links them on its own.

## Layout

```
api/      Go module: cmd/fuda (web server) and cmd/fuda-desktop (Wails app) and internal/ packages; serves the API and the built app
client/   Vite + React + TypeScript app; builds into api/internal/web/dist
```

fuda's own tickets, architecture, decisions, ADRs and glossary are in the fuda board repository, not here.

The Guide pages shown inside the app are in `api/internal/guide/pages/`: they are for people who
use fuda on their repository, not for people changing fuda.

## Rules for code

- **Names explain the code.** No comments that restate it, and no doc comments on functions or
  methods, in Go or TypeScript. A comment only for a non-obvious why.
- **Behaviour lives in `internal/board`.** HTTP handlers parse, call the service and write.
  Interfaces are declared where they are used.
- **No speculative abstractions.** Build what is needed now; leave the rest in
  `docs/decisions.md` § TBD in the fuda board repository.
- **Repository content never breaks fuda.** Invalid files become Problems; a failed sync keeps the
  last snapshot.
- **Tests where the logic is.** Unit tests for parsing and rules, small fixtures in `testdata/`, no
  mocking framework.
- **Component tests render through `renderWithListing`.** Put `// @vitest-environment jsdom` on the
  first line of the `*.test.tsx` file; other tests stay in fast Node. The helper in
  `client/src/test/render.tsx` seeds the Board listing and cleans up after each test. Open dropdowns
  with `userEvent`. Code that reads `BOARD_PREFIX` needs `history.replaceState`, `vi.resetModules()`
  and a dynamic import (see `WorkspacePicker.test.tsx`).
- **Client:** shadcn components (`components/ui` is generated, don't edit it), Tailwind tokens from
  `index.css`, every view state in the URL.
- **The Makefile uses only `go`, `pnpm` and `docker`.** The macOS targets also use `codesign` and
  `hdiutil`, which ship with macOS and the Xcode Command Line Tools.
- **`if` bodies are braced, with a blank line before and after every `if` and before every
  `return`.** `pnpm lint` enforces it; `oxlint --fix` adds them.
- **Formatting-only commits go in `.git-blame-ignore-revs`.** Run
  `git config blame.ignoreRevsFile .git-blame-ignore-revs` once so `git blame` skips them.

## Pull requests

1. Branch from `main`. One topic per PR.
2. `make check` passes (Go lint + tests, client lint + format + types + tests). The `ci` workflow
   runs it on every PR and on every push to `main`.
3. Docs change in the same work as the behaviour: README and the Guide pages in the PR, and
   `docs/architecture.md` and `docs/decisions.md` (when a decision changes) as a commit in the fuda
   board repository.
4. Commit messages say what changed and why, in the imperative ("Restore the last board from the
   cache at startup").
5. A change to the client is tried in the running app (`make dev`) before the PR, not only by tests.
6. The PR description has a "How to test" section. For desktop changes it names `make desktop-macos`.

`main` is protected: every PR needs an approving review from a maintainer, and review threads must
be resolved before merging.

## License

By contributing you agree that your contribution is licensed under the [MIT License](LICENSE).
