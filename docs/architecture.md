# Architecture

How fuda works today. Everything here is true of the code on `main`; anything decided but not
built is in [decisions.md § TBD](decisions.md#tbd). Change this file in the same PR as the code
it describes.

For how a repo is set up and how teams use the board, see the Guide pages in
[`api/internal/guide/pages`](../api/internal/guide/pages) (also served inside the app).

## The shape

One Go binary serves the JSON API and the embedded React app. One process serves many Boards. A Board is one repository, addressed by path: `/github/<owner>/<repo>/`, `/azure/<org>/<project>/<repo>/`, `/local/<folder>/`.
There is no database: the repository's markdown files are the data, and fuda keeps a parsed
snapshot of them in memory, plus a copy on disk so a restart has something to show.

```mermaid
flowchart LR
  subgraph repo["the repository"]
    T["docs/board/tasks/*.md"]
    A["docs/board/archive/*.md"]
    C["docs/board/stages.md · labels.md · people.md · repos.md<br/>(optional)"]
    D["docs/** (other md + images)"]
  end
  subgraph host["git host or local checkout"]
    DEV["develop: the board"]
    MAIN["main: optional, 'in prod'"]
    PR["open PRs in the code repos: In review"]
  end
  subgraph fuda["fuda (one container)"]
    SRC["source: github · azure · local"]
    SVC["board.Service<br/>sync · snapshot · rules"]
    HTTP["httpapi<br/>API · SPA · GitHub login"]
    DISK[("FUDA_CACHE_DIR<br/>snapshot.gob")]
  end
  UI["browser: React app"]
  repo --> DEV
  DEV --> SRC
  MAIN --> SRC
  PR --> SRC
  SRC --> SVC
  SVC <--> DISK
  SVC --> HTTP --> UI
```

## Packages

Dependencies point one way: `cmd/fuda → httpapi → board → (taskfiles, markdown)`, and
`cmd/fuda → source/* → board` (sources satisfy interfaces `board` declares). Interfaces are
declared where they are used.

| Package | Owns |
|---|---|
| `cmd/fuda` | Reads config, picks the source, wires the service and the HTTP server, graceful shutdown. |
| `internal/config` | `FUDA_*` environment variables into one typed `Config`, validated per source. |
| `internal/board` | All behaviour. `token.go`: the caller's host token travels in the request context. `service.go`: the read side (board, task, search, archive, doc, asset). `sync.go`: when and how the snapshot is rebuilt. `cache.go`: the disk copy. `reviews.go`: open PRs in the code repositories → In review. The rules: `columns.go`, `facets.go`, `people.go`, `references.go`, `rules.go`, `board.go`. `move.go`: the Move flow and the retry loop both writes share. `assign.go`: the Assign flow. Declares `Source`, the optional `ReviewSource` and the optional `Writer`. |
| `internal/taskfiles` | Parses what fuda reads from a repo: task files (frontmatter + body) and the optional `stages.md`, `labels.md`, `people.md`, `repos.md`. Invalid files become Problems, never errors. `edit.go`: the pure edits. Move changes the `status` line (and adds `claimed` once). Assign rewrites the `owner` line in its own style. Both leave every other byte alone. |
| `internal/source/local` | A folder on disk with `docs/board/tasks/` (a code checkout) or `tasks/` (a tasks repository). Reads the working tree for develop, uncommitted edits included, and `git archive` for other branches (code checkouts only). Writes a Task file in place. No PRs. |
| `internal/source/github` | GitHub REST with the caller's token: branch head (conditional request), zipball, open pulls and their files, file contents at a commit, the user's `fuda-` repositories. 401 becomes `ErrUnauthorized`, 403 `ErrForbidden`. |
| `internal/source/azure` | Azure DevOps REST 7.1 with the caller's Entra token: refs, items zip of `/docs`, one file with its object id, active PRs, the organizations and `fuda-` repositories of the user, the Contribute permission check. A write is a push with the branch head as `oldObjectId`; it reports "changed since read" when the file's object id moved, and retries when only the branch moved. |
| `internal/markdown` | goldmark + GFM. Rewrites links and images: task files → the task sheet, docs → the reader, images → `/api/files`, other repo paths → the git host's web UI, missing targets → plain text. Raw HTML stays escaped. |
| `internal/httpapi` | Routes, JSON, taking the login token for a request, request logging, panic recovery, the SPA handler. Thin: parse, call `board`, write. |
| `cmd/fuda-desktop` | The Wails v3 desktop shell: same handler and UI, no listening port. Menu, self-update from GitHub Releases (`update.go`). On macOS it ships as `fuda.app` (`Info.plist`, `icon.icns`); update replaces the binary inside it. |
| `internal/app` | Builds the set of Boards from config: which source serves which Board. Shared by both shells. |
| `internal/login` | Login. `device.go`: device flow for GitHub on web and desktop (Azure on desktop), with the token kept in a `TokenStore` (desktop) or a sealed cookie (`cookies.go`, web). `key.go`: the web cookie key. `session.go`: the Azure web login with PKCE and refresh. Knows nothing about Boards. |
| `internal/keychain` | A `login.TokenStore` in the OS keychain (macOS Keychain, Windows Credential Manager). |
| `internal/web` | `go:embed` of the built client (`dist/`). |
| `internal/guide` | `go:embed` of the Guide pages, rendered with `markdown`. |

The desktop app has the same shape as the server: `cmd/fuda-desktop` serves the handler to the Wails
window in-process instead of listening on a port.

`client/` is the Vite app. It builds into `api/internal/web/dist`, so `go build` embeds it.
`go build` works without it; the SPA handler then answers "client not built".

## Boards

`board.Boards` creates one `board.Service` per Board the first time its path is requested, and keeps
it. Each Service has its own snapshot and cache (`FUDA_CACHE_DIR/<host>/<repo>`).

Every request for a Board calls `Service.Poll` with the caller's token in the context. Poll asks the
host for the branch head first, which also proves the caller may read the repository, then reads
files only if the head differs from the snapshot. A caller the host refuses gets 404 (no such
repository for them), 403 or 401 and never sees the cached snapshot. A Board whose first read fails
is forgotten again, so unknown paths do not pile up in memory.

On GitHub and Azure DevOps the token comes from the person's login for that Board's host. A person can
be logged in to both at once: `FUDA_SOURCE=github,azure` starts one login per host. Local Boards have no
login. Their folders come from `FUDA_LOCAL_PATH` (with `FUDA_SOURCE=local`, which stands alone) or, on desktop,
from "Open folder…"; the desktop app remembers them in `folders.json` in the user config directory. `app.Folders`
names each folder after its directory (a second one with the same name gets `-2`) and lists them all in the
Board dropdown's "Local" group.

## Login

GitHub uses device flow on web and desktop, with only the client id (no client secret, no refresh
token: the App has "Expire user authorization tokens" off). The web server and the desktop app show
the same code screen.

```mermaid
sequenceDiagram
  actor U as Person
  participant F as fuda
  participant G as GitHub
  U->>F: GET /api/github/o/fuda-x/board
  F-->>U: 401
  U->>F: GET /auth/github/login?return=…
  F->>G: device code request (client id)
  F-->>U: code screen, device code sealed in a short cookie
  U->>G: type the code, approve the fuda GitHub App
  U->>F: POST /auth/github/device/poll (repeats)
  F->>G: token request with the device code
  F-->>U: sealed session cookie when approved, then back to the Board
```

On the web the session cookie (`fuda_github`) and the pending-login cookie (`fuda_github_device`)
are AES-GCM sealed and HTTP-only, `SameSite=Lax`. The key is made on first start and saved in
`cookie.key` in `FUDA_CACHE_DIR` with mode 0600, so the cache dir must persist or every restart
logs everyone out (the image declares it as a volume). Cookies are `Secure` except on `localhost`,
`127.0.0.1` and `::1`; on a plain-HTTP LAN address the browser drops them and login does not work.
`FUDA_COOKIE_SECRET`, `FUDA_BASE_URL` and `FUDA_GITHUB_CLIENT_SECRET` are not read for GitHub; fuda
logs one warning if they are set. If the API answers 401, the client goes to `/auth/<host>/login`,
so one expired host never logs the person out of another. Each host has its own cookie
(`fuda_github`, `fuda_azure`), so logging out of one leaves the other alone.

On desktop `GET /auth/github/login` also opens GitHub's device page in the system browser. The
token goes to the OS keychain; nothing is written to disk. Each host has its own keychain entry
(`fuda` / `github`, `fuda` / `azure`). The app serves every host whose client id is set
(`FUDA_GITHUB_CLIENT_ID`, `FUDA_AZURE_CLIENT_ID`, or the built-in ones).

### Azure DevOps on the web

Azure keeps the redirect flow with PKCE and a client secret. `/auth/azure/login` redirects to
Microsoft, the callback exchanges the code, and the session cookie (`fuda_azure`) is sealed with a
key derived from `FUDA_COOKIE_SECRET`. It holds the access token, the refresh token and the expiry.
A token that expires within a minute is refreshed; one refresh at a time runs and its result is
remembered for a minute for requests still carrying the old cookie. If refresh fails the cookie is
cleared and the API answers 401. On desktop Azure uses device flow with the keychain, and refreshes
a token that expires within a minute.

## The snapshot

`board.Service` holds an `atomic.Pointer` to an immutable snapshot. Reads never lock; a sync
builds a new snapshot and swaps it in. A snapshot contains:

- the parsed develop result (tasks, archive, config, problems) and, when watched, main;
- the built `Board` (columns, cards, facets, problems), computed once per sync;
- docs (`*.md`) and images under `docs/`, kept as bytes (images up to 5 MB each);
- the open-PR result (which open PRs name each task: In review);
- the head SHA per branch it was built from.

Task bodies and docs are rendered to HTML per request; the board itself is pre-built.

## Sync

```mermaid
flowchart TD
  START["process start"] --> R["restore snapshot.gob<br/>from FUDA_CACHE_DIR, if any"] --> REQ
  REQ["every Board request<br/>(the client polls about every 5 s)"] --> H["Poll: head SHA of develop with the caller's token"]
  H --> SAME{"same head as<br/>the snapshot?"}
  SAME -- yes --> DUE{"PRs last read more than<br/>FUDA_SYNC_COOLDOWN ago?"}
  DUE -- yes --> BG["sync in the background"] --> S
  DUE -- no --> ANS["answer from the snapshot"]
  SAME -- no --> S["sync now: singleflight, 2 min limit"]
  B["Sync button: POST …/sync"] --> BC{"inside the cooldown?"}
  BC -- yes --> TMR["429"]
  BC -- no --> S
  S --> HM["head SHA of develop<br/>(+ main if FUDA_WATCH_MAIN)"] --> F["download docs/ of each branch when a head moved<br/>save snapshot.gob<br/>parse"] --> PRS
  PRS["open PRs into develop<br/>→ task files they change<br/>→ which ones they deliver"] --> SWAP["build the board, swap the snapshot"]
```

- **Failures keep the last snapshot.** A failed head or file read changes nothing and is reported
  as `sync.lastError` in `…/board`. If only the PR list fails, the new files are used without
  In review, and the error is reported. A refusal from the host (login or access) is returned to
  that caller only and is not recorded.
- **After a restart** the cached copy is restored. The first request checks the head with the
  caller's token before it is served. Without a cache, `…/board` answers 503 until a sync succeeds.
- **No webhooks.** Change reaches everyone through polling, about 5 to 10 seconds after a push.

## How a card gets its column

```mermaid
flowchart TD
  F["task file in docs/board/tasks/"] --> V{"valid?"}
  V -- no --> PB["Problems"]
  V -- yes --> PRQ{"an open PR names it,<br/>and a stage has when: pr-open?"}
  PRQ -- yes --> IR["In review"]
  PRQ -- no --> ST{"stages.md?"}
  ST -- yes --> M{"status listed in a stage?"}
  M -- yes --> COL["that stage's column"]
  M -- no --> UNK["a column named after the status,<br/>flagged unknown, at the end"]
  ST -- no --> DEF["default stages: Backlog, In progress, In review,<br/>Merged, Testing, Validated; other statuses flagged unknown"]
```

- **Names:** the task's id appears, as a whole word and in any case, in the title or branch name
  of an open PR in a repository listed in `repos.md`, on any base branch. One PR can name several
  tasks. A card held in In review this way is locked: a Move is refused. Closing the PR returns the
  card to its `status` Stage; fuda never sets `merged`.
- **A repository the caller cannot see** is skipped, so the Board loads without its In review. Any
  other failure keeps the board, drops In review and is reported as the sync error.
- **Statuses compare** case-insensitively with whitespace collapsed.
- **main** never moves a card. When watched, a card whose main copy is merged, testing or validated
  gets the "in prod" badge.
- **Archived tasks** (`docs/board/archive/`) are not on the board; `/api/archive` lists them.

### Validation (what becomes a Problem)

- No frontmatter, unclosed frontmatter, invalid YAML, frontmatter that isn't a mapping.
- Missing `id`, `title` or `status`; a field with the wrong shape.
- A filename that doesn't start with the id.
- `status: done` outside the archive.
- A duplicate id (the first file by path wins).
- A task file in a sub-folder of `tasks/` or `archive/`.
- An invalid `stages.md`, `labels.md`, `people.md` or `repos.md`: the file is ignored and the board is derived
  from the tasks instead.

### Facets (what the filters offer)

- **People:** from `people.md` (names + aliases) when present; otherwise every owner and tester,
  matched case-insensitively, with a unique first name folded into the full name.
- **Labels:** `group:value`. `theme` reads as `epic`; a bare value goes to the `label` group. From
  `labels.md` when present (only listed values are offered), otherwise from the tasks, with
  default `type` values added.
- **Label colours:** `labels.md` `colors:` first, then fixed colours for the default types, then
  a palette in order per group.
- **Statuses, testers, id prefixes:** from the tasks.

## End to end

```mermaid
sequenceDiagram
  actor Dev
  actor Tester
  participant Develop as develop
  participant Main as main (optional)
  participant Fuda as fuda
  Dev->>Develop: claim SS-12 (status in progress, owner, claimed)
  Develop-->>Fuda: poll sees a new head
  Note over Fuda: SS-12 in In progress
  Dev->>Fuda: opens a PR whose title or branch names SS-12
  Note over Fuda: SS-12 in In review, PR #n
  Dev->>Develop: PR completes
  Develop-->>Fuda: poll sees a new head
  Note over Fuda: SS-12 in Merged
  Tester->>Develop: status testing, then validated
  Note over Fuda: Testing, then Validated
  Develop->>Main: promotion
  Main-->>Fuda: poll sees a new head
  Note over Fuda: "in prod" badge
  Dev->>Develop: move to archive/, status done
  Note over Fuda: off the board, in the Archive view
```

The statuses and who changes them are in the Guide's workflow page.

## Read-only Boards

`board.Service.CanWrite` asks the source whether the caller may write (an optional `Access`
interface; GitHub reads `permissions.push` of the repository). The answer is kept for a minute per
token. Move and Assign return `ErrForbidden` for a read-only account, and `GET /board` sets
`readOnly`, so the client shows a note and no drag handles.

## Move

A Move is one commit to one Task file, made with the caller's own token (so the host shows them as
the author). `board.Service.Move` does this:

```mermaid
flowchart TD
  R["POST …/tasks/{id}/move<br/>seen status, target column"] --> C{"archived, or a<br/>PR-derived column?"}
  C -- yes --> F["403"]
  C -- no --> READ["read the file and its version"]
  READ --> SAME{"status still<br/>equals seen?"}
  SAME -- no --> L["409: who moved it, and to where"]
  SAME -- yes --> EDIT["taskfiles.ApplyMove: status line,<br/>claimed once"] --> W["write, compare-and-swap<br/>on the file version"]
  W -- changed since read --> N{"4 tries used?"}
  N -- no --> READ
  N -- yes --> L2["409: the task keeps changing"]
  W -- ok --> SYNC["sync, then 204"]
```

`claimed` is set to today only when the Task leaves the first column, has no `claimed` value, and
moves to another column. fuda never changes a `claimed` value that exists. The write goes to the
work branch (`develop`). GitHub implements `Writer` with the contents API (the file's blob SHA is
the version; a 409 is "changed since read"). Azure DevOps uses the pushes API. The Local source compares the
file's hash with the one it read, then replaces the file through a temporary file and a rename, so an editor's change
made before the write is never overwritten; it does not commit, and it cannot name who changed the file
("Someone moved this to …").

The client keeps no copy of the board for a Move. While a Move is saving, the pending Move is
overlaid on the polled board (`applyPendingMoves`), which also stops the card being dragged
again. When the request ends, the board is refetched and the overlay is gone, so a failed Move
shows the card back in its real column.

## Assign

An Assign is the same kind of commit on the `owner` field. `board.Service.Assign` checks that the
Task is not archived, then runs the retry loop it shares with Move. The request carries the Owners
the client showed (`seen`). If the current Owners differ from `seen` (compared as sets, after
`people.md` aliases are folded), the first write wins: 409 with "Ben set the owners to Cy just
now". Otherwise `taskfiles.ApplyAssign` rewrites the line: comma text, a block list or a flow
list keeps its style, a new line is comma text after `status`, and an empty set deletes the line.

The client overlays pending Assigns on the polled board and on the Task panel the same way as
Moves (`applyPendingAssigns`), and disables the picker while its save runs.

## HTTP API

Board routes sit under `/api/<host>/<board path>`, for example `/api/github/<owner>/<repo>/board`. An unknown Board, or one the caller cannot read, is 404. No login is 401 (`login required`); a host refusal is 403 (`no access`).

| Route | Returns |
|---|---|
| `GET …/board` | Columns, cards, facets, problems, sync state, title, origin. 503 (same body) until a snapshot exists. |
| `GET …/tasks/{id}` | One card plus custom fields and the rendered body. |
| `GET …/search?q=` | Ids of tasks whose title or body contains the text. |
| `GET …/archive` | Archived tasks as cards. |
| `GET …/docs?path=` | A doc under `docs/`: rendered HTML, title, tasks linking to it. |
| `GET …/files?path=` | An image under `docs/`, with `nosniff` and a sandboxing CSP. |
| `GET /api/guide`, `GET /api/guide/{slug}` | Guide pages. |
| `POST …/sync` | 202, or 429 inside the cooldown. |
| `POST …/tasks/{id}/move` | Body `{"seen": "<status the client showed>", "column": "<column id>"}`, `application/json`. 204 when committed. 409 with a message when `status` changed first. 403 for archived Tasks, PR-derived Stages and sources that cannot write. 401 when the login expired. |
| `POST …/tasks/{id}/assign` | Body `{"seen": ["<owners the client showed>"], "owners": ["<short names>"]}`, `application/json`. 204 when committed. 409 with a message when the owners changed first. 403 for archived Tasks and sources that cannot write. 401 when the login expired. |
| `GET /api/boards` | `{boards, hosts}`. `boards` are the Boards the caller can open on every host they are logged in to: `host`, `repo`, `path`, `title`. `hosts` has one entry per login host: `host`, `loggedIn`, `login` (where to log in), and `error` when that host's list failed. Always 200: a host without a login only shows `loggedIn: false`. |
| `GET /auth/<host>/login?return=`, `GET /auth/<host>/callback`, `POST /auth/<host>/logout` | The login flow of `github` or `azure` (only when `FUDA_SOURCE` includes it). `return` must be a path on this site. Logout clears only that host. |
| `POST /auth/logout` | Clears every host's login. 204. |
| `GET /healthz` | 200. |
| anything else | The SPA. Hashed files under `/assets/` are cached for a year; `index.html` is `no-cache`. |

## Client

React 19, TypeScript, Vite, TanStack Router (file routes, typed search params) and TanStack
Query, Tailwind 4 with shadcn components, mermaid loaded lazily.

- **Routes:** `/` board (or list view), `/insights`, `/archive`, `/docs/$`, `/guide/$`. Outside a
  Board, `/` lists the caller's Boards, or opens the only one. One top-bar button shows the current Board (host icon, name, branch) and, as a dropdown grouped by host
  then "Local", switches Boards with a full page load; its current row has "Open repo". An account menu next
  to the theme toggle lists each host with Log in or Log out, and "Log out of all". Any 401 sends the browser to the login, which returns to the same
  page. The board refetches every 5 seconds; when its head moves, tasks, archive and docs refetch.
- **State lives in the URL:** every filter, the sort, the view, the open task (`?task=`) and the
  Insights range, so a link reproduces the view.
- **Filtering, sorting and Insights run in the browser** over `…/board`. Only "also in task
  text" calls `/api/search`.
- **Components** are split into `atoms`, `molecules`, `organisms`; `components/ui` is generated
  by shadcn and not edited by hand.
- **Light and dark themes** come from CSS tokens in `index.css`.

## Configuration

Every setting is an environment variable; the list with defaults is in the Guide's self-host page
and in [`.env.example`](../.env.example). Fixed in code: the board branch `develop`, the prod
branch `main`, the board folder `docs/board/`, the docs root `docs/`.

## Deployment

The `Dockerfile` builds the client (Node), then a static Go binary, into a distroless image that
runs as non-root with a `/data` volume for the cache. One container serves every Board; it gets its
own environment.
