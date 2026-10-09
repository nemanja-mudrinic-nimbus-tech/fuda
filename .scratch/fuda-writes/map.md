# Map: fuda writes Moves and Assigns

Label: wayfinder:map

## Destination

A settled design for how a Move or an Assign made in fuda (web or desktop) gets into git and reaches every other user within about 10 seconds, on GitHub and Azure DevOps, at $0. Recorded as ADRs and a proposal in `docs/decisions.md` § TBD, ready for `/to-spec`.

## Notes

- Domain: fuda, this repo. Read `docs/architecture.md`, `docs/decisions.md` and `CONTEXT.md` first. Use the glossary terms (Task, Move, Assign, Owner, Stage).
- Skills: grilling tickets call `grilling` and `domain-modeling`. Research tickets call `research`.
- Standing preferences: plan, don't build. $0 hosting; free tiers of services are fine. The user prefers short, plain answers.
- Settled while charting (2026-10-08):
  - Task markdown files in the repository stay the only source of truth. Developers and agents keep editing them in git.
  - No extra realtime service and no database: git only. (Rejected: a Cloudflare Durable Object relay; a database as the truth.)
  - Changed 2026-10-08: Tasks live in their own tasks repository, mounted in the code repository as a submodule. This replaced the planned board branch. See "Which branch holds a Task's current Status".
  - The app only Moves and Assigns. Tasks are created and edited in git.
  - Others see a change within 5 to 10 seconds; polling is fine.
  - Sources: GitHub and Azure DevOps. The app talks to the host's web API, not to a local `git` program.
  - One Go core, two shells: the existing web server, and a Wails v3 desktop app that mounts the same Go HTTP handler and React UI. On the web, tokens stay on the server; on desktop, in the OS keychain.
  - Scale: 1 to 3 repositories, under 20 people, some of them non-developers. Everyone has a host account.
- Added 2026-10-08: the web shell stays supported, but fuda picks no host. Hosting it is optional, for anyone willing to pay; it cannot be kept free. The $0 rule applies to the desktop app and the git hosts. See "Where the web shell is hosted at $0".
- This reverses "Read-only" in `docs/decisions.md`. The final ADRs must say so.

## Decisions so far

- [User login and tokens for GitHub and Azure DevOps](issues/03-login-and-tokens.md): GitHub App user tokens and Entra ID are free; Azure DevOps write access is free for only 5 users; private-repo branch protection is paid, so fuda guards the branch.
- [Writing to a board branch through the GitHub and Azure DevOps APIs](issues/01-board-branch-writes.md): one compare-and-swap commit per change on both hosts; fine within limits with per-user tokens; use a normal branch, custom refs unproven.
- [What a Move and an Assign change in a Task file](issues/04-what-a-move-writes.md): Move sets `status` (and `claimed` once); Assign keeps the file's `owner` style with `people.md` short names; no card order; one-line edits.
- [Cost of polling for changes on GitHub and Azure DevOps](issues/02-change-polling-cost.md): 5 s polling is cheap on both hosts (GitHub 304s are free); no free push signal.
- [Which branch holds a Task's current Status](issues/05-status-source-of-truth.md): no board branch; Tasks move to their own repository (one branch), mounted as a submodule in the code repository. Needs an ADR.
- [How board-branch changes reach develop](issues/06-board-branch-to-develop.md): not needed after the tasks-repository decision.
- [Whose identity writes the commit, and how each shell gets the token](issues/08-who-writes-the-commit.md): everyone writes as themselves (Azure seats may cost); device code on desktop; encrypted cookie on web, keychain on desktop; login first, repo access checked, read-only if no write.
- [Two changes to the same Task at once](issues/07-concurrent-edits.md): compare-and-swap, never force; other-line changes retry silently; same field changed first: first wins, card snaps back with a message.
- [How developers and agents keep the tasks submodule up to date](issues/09-submodule-pointer.md): `branch = main` plus `update --remote`, never bump the pointer; manual edits inside `docs/board` stay supported; documented in CONTRIBUTING and an agent rule.
- [How a user chooses which board to look at](issues/10-choosing-a-board.md): many Boards per fuda; a Board is a `fuda-` repo found from the user's login; URL `/<owner>/<repo>/` with the prefix kept; picker in the top bar; reads with the user's token, no server token.
- [How the board shows a Move that is not yet confirmed](issues/11-unconfirmed-move.md): optimistic move with a saving dot and no re-drag; on failure the card goes back with a reason; warn on leaving mid-save.
- [Where the web shell is hosted at $0](issues/17-web-hosting.md): no managed free host has disk plus no sleep; Oracle Always Free VM with Caddy, or Fly.io at about $3 a month.
- [Shipping the Wails desktop app at $0](issues/18-desktop-distribution.md): unsigned works but warns; macOS needs $99/yr to avoid it, Windows signing free via SignPath (open source) or $9.99/mo Azure; Wails updater on free GitHub Releases.
- [How Moves interact with In review from PRs and the archive](issues/12-moves-and-derived-status.md): In review stays PR-derived (code repo PRs); such cards are locked; archived Tasks are read-only in the app.
- [What the local source becomes when fuda writes](issues/13-local-source.md): stays as a Local Board; folder picker on desktop, env on web; Moves and Assigns write the file on disk, the person commits; polls the working tree; same conflict rule as the host.
- [Moving existing Task files into a tasks repository](issues/14-migrate-task-files.md): keep history with `git subtree split`; a manual checklist in CONTRIBUTING, no command; no link redirects.
- [A user logged in to GitHub and Azure DevOps at once](issues/15-two-hosts.md): both hosts at once in one picker; host first in URL (`/github/...`, `/azure/<org>/<project>/...`); login and logout per host plus "Log out of all".
- [A fuda- repository with no Task files or no board config](issues/16-empty-board-repo.md): empty board with default Stages and a hint; new Board = create an empty `fuda-` repo on the host; invalid config stays a Problem with defaults.
- [How a code PR names its Task](issues/19-pr-names-its-task.md): Task id in PR title or branch (all ids count, any base); code repos listed in board config; user's token; a person Moves to Merged.

## Not yet specified

- Nothing. ADRs 0001 to 0004 and the TBD proposal "Moves and Assigns from the app" in `docs/decisions.md` are written. Next: `/to-spec`.

## Out of scope

- Creating and editing Task text in the app: ruled out while charting; Tasks are written in git.
- Updates faster than about 2 seconds: not needed; would force a realtime service.
- GitLab and other hosts.
- A Task-creation skill for other people's agents: creating Tasks is not Moves or Assigns; a separate effort (leaning: writes through git). From "How developers and agents keep the tasks submodule up to date".
- Picking or documenting a specific web host: the web shell is self-hosted by whoever wants it, at their cost. From "Where the web shell is hosted at $0".
