# What the local source becomes when fuda writes

Type: grilling
Status: resolved
Blocked by: 

## Question

The local source reads a checkout on disk. Now Boards are `fuda-` repos found from the user's login, and writes go through the host API. Does the local source stay (read-only dev mode, or writes to disk), point at a tasks-repository checkout, or go away?

## Answer

Resolved 2026-10-08. The local source stays, as a **Local Board**.

- Desktop: "Open folder…" in the board picker opens the OS file picker; fuda remembers opened folders. Web server and `make dev`: `FUDA_LOCAL_PATH`.
- The folder can be a tasks-repo checkout or a code-repo checkout: fuda looks for `tasks/` first, then `docs/board/tasks/`.
- Local Boards show in the board picker in a "Local" group next to host Boards. URL `/local/<folder name>/`. No login.
- Reads only the working tree (uncommitted edits included). No branches, no `git archive`, no PRs, so no In review.
- Move and Assign write the file on disk. The person commits. This replaces the "read-only dev mode" first chosen.
- Changes on disk are found by polling the folder every few seconds (no file watcher).
- Concurrent edits: same rule as the host (see "Two changes to the same Task at once"): fuda checks the file is unchanged since read; other-line change → rewrite silently; same field changed → the file wins, the card snaps back with a message.
