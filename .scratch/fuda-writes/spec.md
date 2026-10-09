# Spec: Moves and Assigns from fuda

Status: ready-for-agent

Design: the map `.scratch/fuda-writes/map.md`, ADRs 0001 to 0004 in `docs/adr/`, and the TBD
proposal "Moves and Assigns from the app" in `docs/decisions.md`.

## Problem Statement

fuda is read-only. To move a Task to another Stage or to change its Owner, a person must edit a
markdown file in git and push it. Developers and agents can do this. Other people on the team
cannot, so they ask a developer, or the board goes out of date. The Tasks also live inside the code
repository, so anyone who needs only the board also needs access to the code.

## Solution

People log in to fuda with their GitHub or Azure DevOps account. They drag a card to Move it and
pick people to Assign it. fuda writes a one-line commit to the Task file as that person. Everyone
else sees the change within about 5 to 10 seconds. Tasks live in their own `fuda-` tasks
repository, which the code repository mounts as a submodule. fuda runs as the existing web server
(self-hosted) or as a free desktop app. Git stays the only source of truth: developers and agents
keep editing Task files as before.

## User Stories

Boards and login

1. As a team member, I want to log in with my GitHub account, so that fuda knows who I am without a separate password.
2. As a team member, I want to log in with my Azure DevOps (Microsoft) account, so that I can use fuda on Azure-hosted repositories.
3. As a team member with work on both hosts, I want to be logged in to GitHub and Azure DevOps at once, so that I see all my Boards in one place.
4. As a team member, I want to log out of one host only, so that my other host stays logged in.
5. As a team member, I want a "Log out of all" button, so that I can clear every login at once.
6. As a team member, I want fuda to find my Boards from my login, so that nobody has to configure a list of Boards.
7. As a team member, I want every repository whose name starts with `fuda-` to show as a Board, so that adding a Board is just creating a repository.
8. As a team member, I want a Board picker in the top bar, so that I can switch between Boards quickly.
9. As a team member with exactly one Board, I want `/` to open it, so that I skip an extra click.
10. As a team member, I want the URL to name the host and repository (`/github/<owner>/<repo>/`, `/azure/<org>/<project>/<repo>/`), so that I can share a link to a Board.
11. As a team member whose host login expired, I want only that host to send me back to login, so that my Boards on the other host keep working.
12. As a team member without access to a repository, I want a clear message, so that I know why I cannot see the Board.
13. As a team member who can read but not write a repository, I want a read-only board with a note saying why, so that I am not confused when cards don't drag.
14. As an org admin on GitHub, I want to install the fuda GitHub App once per org on `fuda-` repositories, so that fuda sees only those repositories.
15. As a new team, I want an empty `fuda-` repository to show as an empty Board with default Stages and a hint, so that I can start from nothing.

Move

16. As a team member, I want to drag a card to another Stage, so that I can change its Status without git.
17. As a team member, I want the card to move at once when I drop it, so that the board feels fast.
18. As a team member, I want a small saving dot on the card while it saves, so that I know the change is not yet in git.
19. As a team member, I want the card to stay still until its save ends, so that I don't stack changes on an unsaved one.
20. As a team member, I want the card to go back with a short reason when a save fails, so that I know my change did not happen.
21. As a team member whose login expired during a save, I want a "log in again" link, so that I can fix it fast.
22. As a team member, I want the browser to warn me if I leave while a save is running, so that I don't lose a change by accident.
23. As a team member, I want fuda to set `claimed` to today the first time a Task leaves the first Stage, so that the claim date is recorded without extra work.
24. As a team member, I want `claimed` never changed again once set, also when a Task moves back, so that the first claim date is kept.
25. As a developer, I want a Move to change only the `status` line (and add `claimed` once), so that the diff is tiny and the rest of my file stays byte-for-byte.

Assign

26. As a team member, I want to pick the Owners of a Task, so that everyone sees who works on it.
27. As a team member, I want to pick Owners by their short names from `people.md`, so that names match the rest of the board.
28. As a team member on a repository without `people.md`, I want to pick from the names already on the board, so that Assign still works.
29. As a developer, I want Assign to keep the file's `owner` style (comma text or YAML list), so that my files stay consistent.
30. As a team member, I want removing every Owner to delete the `owner` line, so that the file stays clean.

Commits and conflicts

31. As a team member, I want my Move or Assign committed as me, so that history shows who did it.
32. As a developer, I want fuda never to force-push, so that none of my commits are lost.
33. As a team member, I want my change to go through silently when someone changed other lines first, so that I am not bothered by unrelated edits.
34. As a team member, I want the first change to win when someone changed the same field first, so that fuda never overwrites a value I never saw.
35. As a team member who lost a conflict, I want a short message such as "Ben moved this to Review just now", so that I understand what happened.
36. As a team member, I want to see other people's changes within about 5 to 10 seconds, so that the board is current without reloading.

In review and the archive

37. As a team lead, I want to list the code repositories of a Board in its board config, so that fuda knows where to look for PRs.
38. As a developer, I want a PR whose title or branch name contains a Task id to put that Task In review, so that I don't edit the Task file for review.
39. As a developer, I want a PR that names two Task ids to put both In review, so that one PR can deliver several Tasks.
40. As a developer, I want PRs on any base branch to count, so that repositories using main or develop both work.
41. As a team member, I want a card in In review because of an open PR to be locked, so that the board does not fight the PR.
42. As a team member, I want no card to be dropped into a PR-derived Stage by hand, so that In review always means an open PR.
43. As a team member, I want the card to return to its `status` Stage when the PR closes, so that I can then Move it to Merged.
44. As a team member who cannot see a code repository, I want to still use the Board, without In review from that repository, so that missing code access does not block me.
45. As a team member, I want archived Tasks to be read-only in the app, so that archiving stays a file move in git.

Local Board

46. As a developer, I want to open a local folder as a Board ("Open folder…" on desktop, `FUDA_LOCAL_PATH` on the web), so that I can use fuda offline or on my working tree.
47. As a developer, I want fuda to find `tasks/` or `docs/board/tasks/` in that folder, so that both a tasks-repo checkout and a code-repo checkout work.
48. As a developer, I want Moves and Assigns on a Local Board to write the file on disk, so that I commit them myself.
49. As a developer, I want a Local Board to pick up my editor changes within a few seconds, so that the board matches the disk.
50. As a developer, I want the same conflict rule on a Local Board, so that fuda never overwrites a field I just changed on disk.
51. As a developer, I want desktop fuda to remember folders I opened, so that I don't pick them every time.

Shells

52. As a non-developer, I want a desktop app, so that I can use fuda without anyone hosting a server.
53. As a desktop user, I want my tokens kept in the OS keychain, so that they are stored safely.
54. As a desktop user, I want the app to update itself from GitHub Releases, so that I stay on the newest version.
55. As someone who self-hosts fuda, I want the web server to keep tokens only in an encrypted cookie, so that the server stores no user secrets.
56. As someone who self-hosts fuda, I want no shared basic-auth password and no server host token, so that access is decided by the host alone.

Migration and docs

57. As a maintainer, I want a checklist in `CONTRIBUTING.md` to move `docs/board` into a `fuda-` repository with its history, so that migration is safe and repeatable.
58. As a developer, I want `CONTRIBUTING.md` to explain `git submodule update --remote docs/board`, so that I see the newest Tasks.
59. As an agent, I want a rule in the code repository's agent instructions to update the submodule before reading Tasks and to commit Task edits inside `docs/board`, so that I don't act on old Tasks.
60. As a developer, I want manual Task edits in git to keep working, so that fuda is a convenience and not the only way.

## Implementation Decisions

- **Domain words.** Task, Move, Assign, Owner, Stage, Board, Local Board, as in `CONTEXT.md`. A Board is one tasks repository. Add Board and Local Board to `CONTEXT.md` if missing.
- **Many Boards per process.** Today the service serves one repository from env config. It becomes a set of Boards, each identified by host plus repository (or local folder), created on demand for the logged-in user. Each Board keeps its own snapshot. The single-repository env settings and the cache restart path apply per Board.
- **Source interface grows a write.** The board package declares what it needs: list Boards for a user, read the default branch (head and files), write one file as a compare-and-swap on the file's previous version (or the branch head), and list open PRs in a repository. Write returns a distinct "changed since read" error. GitHub (contents API with the file SHA), Azure DevOps (pushes API with the old object id) and Local (compare the file on disk) implement it. Each call takes the user's token.
- **The edit is a pure function in taskfiles.** Input: file bytes plus a Move (new status, today's date, first Stage name) or an Assign (Owner names). Output: new bytes. Only the changed line changes. A new `claimed` or `owner` line goes right after `status`. The `owner` style follows the file; a new one is comma text; an empty set deletes the line.
- **The Move/Assign flow lives in board.** Read the current file, refuse if the Task is archived, PR-locked or the target Stage is PR-derived, apply the edit, write with compare-and-swap. On "changed since read": re-read; if the same field (`status` for Move, `owner` for Assign) now differs from what the user saw, return a conflict with who and what; otherwise apply again and retry a few times. Never force.
- **Request carries what the user saw.** Move and Assign requests include the Task id and the field value the client showed, so the server can tell "same field changed first".
- **HTTP API.** New endpoints to list the user's Boards, Move a Task and Assign a Task, scoped by the Board path (`/github/...`, `/azure/...`, `/local/...`). Errors: conflict (with a message), forbidden (read-only), unauthorized (login again per host). Handlers stay thin.
- **Auth.** GitHub App user tokens: web flow with PKCE on the web shell, device flow on desktop. Azure DevOps: Microsoft Entra ID app, confidential client on the web shell, MSAL device code on desktop. One token per host: web stores each in its own encrypted cookie; desktop in the OS keychain. Refresh tokens are used before sending a user back to login. At login and when opening a Board, fuda checks read access and write access; no write access means a read-only Board.
- **Removed.** The server's host token, `FUDA_AUTH_USER`/`FUDA_AUTH_PASSWORD`, the single-repository config, and matching PRs by their version of the Task file. Webhooks no longer drive syncs for user-scoped reads; the client polls.
- **Polling.** The client polls the Board about every 5 seconds. The server compares the head SHA first (GitHub conditional requests return 304 at no rate cost) and reads files only when the head changed.
- **In review.** The board config gains a list of code repositories. fuda lists open PRs in each with the user's token, any base branch, and finds Task ids in the PR title and branch name. Every id found counts. A PR-matched Task shows in the PR-derived Stage and is locked. When the PR closes, the Task shows in its `status` Stage. fuda never sets `merged`.
- **Local Board.** Reads the working tree only, polls the folder every few seconds, writes the file on disk with the same compare rule. No PRs, no In review.
- **Client.** Board picker in the top bar, host-first routes, drag-and-drop Move with optimistic state, saving dot, no re-drag while saving, rollback with message, leave-page warning, an Owner picker for Assign, read-only and locked states, per-host login and logout, "Log out of all".
- **Desktop shell.** A Wails v3 app that mounts the same Go HTTP handler and embedded React UI. It adds the folder picker, the keychain store and the updater (GitHub Releases). The web server stays a separate entry point over the same core.
- **Empty and invalid repositories.** No Task files: an empty Board with default Stages and the hint "No Tasks yet: add files in `tasks/`". Invalid board config stays a Problem and defaults apply.
- **Docs in the same change.** Move "Moves and Assigns from the app" from TBD into the settled sections of `docs/decisions.md`, replacing "Read-only", "develop is the board" and the content-matched In review rule. Mark ADRs 0001 to 0004 accepted. Update `docs/architecture.md`, README, the Guide pages and `CONTRIBUTING.md` (migration checklist, submodule section, new Board step).

## Testing Decisions

- A good test drives external behaviour: call Move or Assign and check the resulting file bytes, the result or error, and what the board shows. No checks on internal calls.
- **Main seam: the board service with an in-memory fake host.** The fake holds files per branch, supports compare-and-swap writes, can inject a concurrent change between read and write, and serves open PRs. Tests cover: Move and Assign happy paths, `claimed` set once, other-line change retried silently, same-field change returns a conflict with who and what, never overwriting, read-only accounts refused, archived and PR-locked Tasks refused, drop into a PR-derived Stage refused, PR title and branch id matching (several ids, any base), the card returning after a PR closes, an empty repository, and the Local Board compare rule.
- **Pure edit tests in taskfiles.** Table tests with small fixtures in `testdata/`: status change, `claimed` insert, comma and YAML `owner` styles, removing every Owner, untouched bytes elsewhere.
- **HTTP smoke test.** Extend the existing one to cover a Move through the handler with the fake host, including conflict and forbidden status codes.
- **Client.** Unit tests in `lib` for pure logic only (optimistic state and rollback, Task id matching if done client-side), like the existing filters and insights tests.
- Not unit tested: the GitHub and Azure DevOps HTTP clients, OAuth flows and the Wails shell. They are thin glue; check them by hand against a real test repository.
- Prior art: `board_test.go`, `taskfiles_test.go`, `api_test.go`, `filters.test.ts`, `insights.test.ts`.

## Out of Scope

- Creating Tasks or editing Task text in the app.
- Updates faster than about 2 seconds, or any realtime service.
- GitLab and other hosts.
- A Task-creation skill for other people's agents.
- Picking or documenting a specific web host.
- A "New Board" button, a migration command, or link redirects.
- Storing card order inside a Stage.
- fuda setting `merged` on its own.

## Further Notes

- Azure DevOps write access is free for only 5 users; more writers may need paid seats. Accepted.
- Unverified from research: loopback ports for GitHub Apps, whether users can consent to `vso.code_write` without an admin, and the Basic seat price. Check these first when building Azure login.
- Wails v3 is still beta.
- Unsigned desktop builds warn on macOS and Windows. Signing is a later choice ($99/yr Apple; SignPath free for open source on Windows).
- This is large. Split it into vertical slices when turning it into tickets (for example: tasks repo and many Boards read-only; login; Move; Assign; In review from code PRs; Local Board writes; desktop shell).
