---
status: accepted
---

# fuda writes Moves and Assigns through git

fuda was read-only: every change was a commit by a person or their agent. We decided that fuda will write two things, a Move (`status`) and an Assign (`owner`), so people who are not developers can use the board. Each write is a one-line commit to the Task file, made through the host's web API (GitHub, Azure DevOps) as a compare-and-swap, never a force-push. The Task files stay the only source of truth. Other users see a change by polling every 5 seconds.

This reverses "Read-only" in `docs/decisions.md`. fuda still does not create or edit Task text.

## Considered Options

- A realtime relay (Cloudflare Durable Object) for instant updates: rejected. It is another service to run and pay for. 5 to 10 seconds is fast enough.
- A database as the truth, synced to git: rejected. Two sources of truth, and agents read git.
- A local `git` program: rejected. The web shell and the desktop app both talk to the host API, so there is one write path.
