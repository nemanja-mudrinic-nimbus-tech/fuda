# 11: Local Board writes

**What to build:** A folder (`FUDA_LOCAL_PATH` on the web) is a Local Board at `/local/<folder>/`, shown in a "Local" group in the picker, with no login. fuda finds `tasks/` or `docs/board/tasks/`, reads the working tree, and polls it every few seconds. Move and Assign write the file on disk; the person commits. Same conflict rule: if the field changed on disk, the file wins and the card goes back with a message.

**Blocked by:** 05: Assign on GitHub

**Status:** ready-for-agent

- [x] Board service tests with a temp folder: Move, Assign, other-line change, same-field conflict
- [x] No In review on a Local Board
- [x] Edits made in an editor show within a few seconds
- [ ] The desktop app has "Open folder…" and remembers opened folders

Spec: `.scratch/fuda-writes/spec.md`.
