# 01: Many Boards with host-first URLs

**What to build:** fuda serves Boards by path, `/github/<owner>/<repo>/`, instead of one repository from env config. The board service keeps one snapshot (and cache) per Board, created on demand. Reads still use today's env token. A prefactor so later tickets are easy; the board looks the same to users.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] Opening `/github/<owner>/<repo>/board` shows that repository's Board
- [ ] Two Boards can be open at once, each with its own snapshot and cache
- [ ] Unknown Board path gives a clear not-found page
- [ ] Existing board and service tests pass; new tests cover two Boards side by side
- [ ] `docs/architecture.md` updated

Spec: `.scratch/fuda-writes/spec.md`.
