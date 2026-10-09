# Decisions

The choices fuda is built on, each with the reason. These are settled: change one only with a PR
that updates this file and the code together. Open items are at the end under [TBD](#tbd), each
with a proposal.

## Product

- **fuda writes Moves and Assigns through git.** A person drags a card and fuda commits one
  changed `status` line (plus `claimed` the first time the Task leaves the first Stage) to the
  Task file as that person. Picking Owners commits one changed `owner` line the same way, through the host's web API, as a compare-and-swap, never a
  force-push. Git stays the only source of truth; people and agents keep editing files. fuda
  never creates Tasks or edits their text. Built for Moves and Assigns on GitHub; see
  [Moves and Assigns from the app](#moves-and-assigns-from-the-app) for the rest and ADRs
  [0001](adr/0001-fuda-writes-moves-and-assigns.md) and
  [0002](adr/0002-tasks-in-their-own-repository.md).
- **Assign writes `owner` in the file's own style.** Comma text stays comma text, a YAML list
  stays a YAML list, and a new `owner` line is comma text right after `status`. Removing every
  Owner deletes the line. The names come from `people.md`, or from the names already on the board
  when there is none. Archived Tasks cannot be assigned.
- **A Move or Assign that loses a conflict goes back.** If other lines of the file changed first, fuda
  applies the edit again, up to four times. If `status` changed first, the first write wins: the
  card goes back with a message such as "Ben moved this to Testing just now". An Assign works the
  same on `owner`: if the Owners changed first, fuda says "Ben set the owners to Cy just
  now". Archived Tasks and the PR-derived Stage cannot be Move targets.
- **develop is the board.** Columns and filters come from develop. Claims are committed to develop
  so everyone sees them. main is optional and only adds the "in prod" badge; it never moves a
  card, because anything on main is on develop too.
- **In review comes from open PRs that name the task.** The board lists its code repositories in
  `repos.md` (`code_repos: [org/app]`). An open PR in any of them, on any base branch, puts every
  task whose id is in its title or branch name In review. Such a card is locked: it cannot be
  dragged, a Move is refused, and nothing is dropped into that Stage by hand. When the PR closes
  the card returns to its `status` Stage; fuda never sets `merged`. A person who cannot see a code
  repository still gets the Board, without that repository's In review. Ids are matched whole,
  case-insensitively: `T-1` does not match `T-12`.
- **The repository decides its board.** Columns come from an optional `stages.md`, label groups
  from `labels.md`, people from `people.md`. Without them fuda derives everything from the tasks,
  so a repo works on day one.
- **Never break on repository content.** An invalid file becomes a Problem shown in the UI and is
  left off the board. An unexpected status gets its own flagged column instead of disappearing.
- **A claimed task stays where its status says.** An owner on a backlog task is not a Problem;
  people move tasks explicitly.

## Architecture

- **One binary, one container; one Board per repository path.** The API and the built client ship together
  (`go:embed`). One process serves many Boards, each with its own snapshot, cache and sync.
- **No database.** The files are the data. A parsed snapshot lives in memory and a copy on disk
  (`FUDA_CACHE_DIR`) so a restart serves the last board at once.
- **The client polls; there are no webhooks.** The browser asks for the Board about every 5
  seconds. Each request first compares the branch head SHA (on GitHub a conditional request, which
  costs no rate limit when nothing changed) and downloads files only when it moved. Pull requests
  are re-read at most every `FUDA_SYNC_COOLDOWN`. Updates reach everyone within about 10 seconds.
- **Each person reads with their own token.** On GitHub they log in with the fuda GitHub App
  (device flow); the token lives in a sealed cookie, so the server stores no user secret. Every
  request checks the head with that person's token, so a cached Board is never shown to someone
  GitHub would refuse. On Azure DevOps an expired token is refreshed; if that fails the person goes
  back to login. See
  [ADR 0003](adr/0003-users-read-and-write-with-their-own-token.md).
- **One device-flow login for GitHub.** GitHub login uses device flow on web and desktop, with
  only the client id. The GitHub App has "Expire user authorization tokens" turned off, so there is
  no refresh and no client secret. Desktop keeps the token in the OS keychain. Web keeps it in a
  sealed HttpOnly cookie; the server makes its own key on first start and saves it as `cookie.key`
  in `FUDA_CACHE_DIR` (mode 0600), so the cache dir must persist: the Docker image declares it as a
  volume, and the self-host Guide page says so. Cookies are `Secure` except on localhost, so a
  plain-HTTP LAN address has no working login. The Sign in button in the app bar starts login; there
  is no automatic login screen. `FUDA_GITHUB_CLIENT_SECRET`, and for GitHub `FUDA_COOKIE_SECRET` and
  `FUDA_BASE_URL`, are ignored with one warning. Azure keeps its redirect login.
- **A Board is a `fuda-` repository the person can read.** The Board dropdown lists them from
  `GET /user/repos`, which only returns repositories the app is installed on. `/` opens the only
  Board or lists them.
- **Plain Go packages by responsibility.** No ports-and-adapters layering. `board` holds the
  behaviour and declares the small interfaces it needs; sources satisfy them implicitly. Handlers
  stay thin.
- **The standard library first.** `net/http` routing with a tiny middleware chain, `log/slog`,
  plain HTTP clients for GitHub and Azure instead of SDKs. Dependencies: env parsing, YAML,
  goldmark, singleflight.
- **Filtering runs in the browser.** The board payload carries frontmatter only, and every filter
  is a URL parameter, so views are shareable and the server stays simple.
- **No shared password and no server host token.** Access to a GitHub Board is decided by GitHub.
  The `local` source has no login until its ticket lands; keep such a server on a private network.
- **Configuration is environment only.** Typed and validated at start; no config files for fuda
  itself.

## Code

- **Names explain the code.** No comments that restate it, and no doc comments on functions or
  methods, in Go or TypeScript. A comment only for a non-obvious why.
- **No speculative abstractions (YAGNI).** Build for the two real repositories, not imagined
  ones. Remove features that don't work rather than leave them half-wired.
- **Tests where the logic is.** Unit tests for parsing and rules, one HTTP smoke test, no mocking
  framework, no end-to-end suite. Small fixtures in `testdata/`.
- **The Makefile uses only `go`, `pnpm` and `docker`,** so anyone can clone and run it.
- **Docs change with the code.** README, Guide pages and these docs are updated in the same PR as
  the behaviour they describe.

## TBD

Decided as a direction, not built. Each has a proposal; a PR that builds one moves it above and
updates [architecture.md](architecture.md).

### Moves and Assigns from the app

- **Settled:** fuda will write a Move (`status`) and an Assign (`owner`), and nothing else, as
  one-line commits through the host's web API, on GitHub and Azure DevOps. Task files stay the
  only source of truth; no database, no realtime service. See ADRs
  [0001](adr/0001-fuda-writes-moves-and-assigns.md),
  [0002](adr/0002-tasks-in-their-own-repository.md),
  [0003](adr/0003-users-read-and-write-with-their-own-token.md) and
  [0004](adr/0004-desktop-app-with-wails.md). "Read-only" is already replaced above. This will
  also replace "develop is the board".
- **Open:** nothing to decide. Ready for a spec.
- **Proposal:**
  - **Boards.** A Board is a tasks repository whose name starts with `fuda-`, mounted in its code
    repository as a submodule at `docs/board`. fuda finds Boards from the user's login and lists
    them in the Board button's dropdown in the top bar. URLs start with the host: `/github/<owner>/<repo>/`,
    `/azure/<org>/<project>/<repo>/`, `/local/<folder>/`. An empty `fuda-` repo is an empty Board
    with default Stages and a hint.
  - **Login (rest).** GitHub device flow (web: sealed cookie; desktop: OS keychain) is built. So is Microsoft
    Entra ID for Azure DevOps: a confidential client on the web, the OAuth device code on desktop
    (plain HTTP, no MSAL library), scope `user_impersonation`. One process serves
    both hosts at once (`FUDA_SOURCE=github,azure`), with one cookie or keychain entry per host,
    per-host logout and "Log out of all". Whether
    people can consent to the Azure DevOps scope without an admin depends on the tenant's consent
    policy; not checked against a real tenant yet. A narrower `vso.code_write` scope is not used.
    Read-only Boards: Azure DevOps asks the Contribute permission of the repository.
  - **Move and Assign (rest).** Built for GitHub and Azure DevOps, including read-only Boards: fuda asks GitHub
    whether the account can push, keeps the answer for a minute, and refuses Move and Assign
    for read-only accounts. The board shows a note and no card drags. Local Boards write the file
    on disk. No card order is stored.
  - **Local Board.** A folder (desktop: "Open folder…"; web: `FUDA_LOCAL_PATH`). Moves and Assigns
    write the file on disk and the person commits. fuda polls the folder; same conflict rule. A folder
    is a Local Board when it has `docs/board/tasks/` or `tasks/`. A Local Board has no In review.
    It cannot tell who changed a file, so a lost conflict says "Someone".
  - **Shells.** One Go core: the web server and a Wails v3 desktop app with the same handler and
    UI. The web shell is self-hosted at the host's own cost. The desktop app logs in to GitHub or Azure DevOps
    (device flow, keychain, self-update from GitHub Releases). "Open folder…" in the File menu opens
    a Local Board; the app remembers the folders.
  - **Migration and docs.** `CONTRIBUTING.md` gets a checklist (create the `fuda-` repo,
    `git subtree split --prefix=docs/board`, push as `main`, replace `docs/board` with the
    submodule) and a short section on `git submodule update --remote docs/board`. The code
    repository's agent rules say to update the submodule before reading Tasks.
  - **Removed.** Nothing left to remove. The server's host token, `FUDA_AUTH_USER`/
    `FUDA_AUTH_PASSWORD`, webhooks and matching PRs by their version of the Task file are gone.

### Signing and notarizing the macOS app

- **Settled:** `fuda.app` ships unsigned. People right-click, Open the first time.
- **Open:** whether to pay for an Apple Developer account.
- **Proposal:** Sign `fuda.app` with a Developer ID certificate and notarize it in the `release`
  workflow (`codesign`, `xcrun notarytool`, `stapler`), with the certificate and App Store Connect
  key as repository secrets. Then the first-run warning goes away. Self-update replaces the inner
  binary, which breaks the signature, so it must move to replacing the whole bundle first.

### History-based dates and insights

- **Settled:** fuda only knows the dates written in frontmatter (`added`, `claimed`, `done`).
  "Time in progress", "merged per week" and an "archive candidate" badge all need the date a
  task's status changed.
- **Open:** how far back to read, and the cost on large repositories.
- **Proposal:** for each task, read the file's commit history once from the git host (GitHub
  commits by path, Azure commits by `itemPath`), find the first commit where each status appears,
  and keep the result in `FUDA_CACHE_DIR` keyed by task and blob SHA, so only changed files are
  read again. Expose the dates per card; the client computes the insights. The archive-candidate
  badge (`FUDA_ARCHIVE_AFTER_DAYS`) returns on top of the same dates.

### CI for pull requests

- **Settled:** contributors open PRs; only maintainers merge, and `main` requires a review.
- **Open:** nothing runs `make check` on a PR yet.
- **Proposal:** a GitHub Actions workflow that installs Go, Node and pnpm and runs `make check`
  on every PR, set as a required status check on `main`.

### More git hosts

- **Settled:** a host is a package under `internal/source` implementing `board.Source` (and
  `board.ReviewSource` for In review).
- **Open:** which host is next.
- **Proposal:** GitLab, when someone needs it: project archive for `docs/`, merge requests and
  their changes for In review.
