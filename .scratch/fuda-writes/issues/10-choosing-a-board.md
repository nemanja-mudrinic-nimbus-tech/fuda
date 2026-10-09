# How a user chooses which board to look at

Type: grilling
Status: resolved
Blocked by: 08

## Question

Today one fuda server serves one repository, set in env (`FUDA_GITHUB_REPO` or `FUDA_AZURE_*`); the web address picks the board. With per-user login, one account may see several tasks repositories. Does one server (and one desktop app) serve several boards? If so: who lists the boards (server config, or whatever the user's account can see), how a user switches between them, and what the URL looks like. If not: how the desktop app knows which repository to open.

## Answer

Resolved 2026-10-08 (grilling).

- One fuda (web server or desktop app) shows many Boards. A Board is one tasks repository.
- A repository is a Board when its name starts with `fuda-`. No config list of Boards.
- fuda finds Boards from the user's login:
  - GitHub: `GET /user/repos` with the GitHub App user token, so only repos the App is installed on. An org admin installs the App once per org, ideally on `fuda-` repos only. (Rejected: an OAuth App; it needs the wide `repo` scope and is often blocked by org policy anyway.)
  - Azure DevOps: list the user's organizations, then the repositories in each.
- URL: `/<owner>/<repo>/...` with the full repo name, prefix kept. Example: `acme/fuda-web` → `/acme/fuda-web/board`. (Stripping the prefix to `/acme/web` read like an app route.)
- A Board picker in the top bar. `/` lists the user's Boards; with exactly one, `/` opens it.
- Local source: scans a folder for `fuda-*` subfolders.
- Boards are read with the user's own token. The web server token goes away. This changes "Whose identity writes the commit" (it said the web shell reads with its server token). Polling at 5 s is well inside each user's rate limit.

## Comments

- 2026-10-08: Changed by "A user logged in to GitHub and Azure DevOps at once" (15): the URL now starts with the host, `/github/<owner>/<repo>/` or `/azure/<org>/<project>/<repo>/`.
