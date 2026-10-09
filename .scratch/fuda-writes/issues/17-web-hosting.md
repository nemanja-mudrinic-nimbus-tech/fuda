# Where the web shell is hosted at $0

Type: research
Status: resolved
Blocked by: 

## Question

Which free hosts can run the fuda Go binary (one container, HTTPS, a secret for cookie encryption, small disk for FUDA_CACHE_DIR, no sleep or a fast wake)? Compare free tiers (e.g. Fly.io, Render, Koyeb, Google Cloud Run, Oracle Cloud free VM) on limits, cold start, and whether a card is needed.

## Answer

No managed free host gives HTTPS, a secret, a persistent disk and no sleep together. Free and always-on: an Oracle Cloud Always Free Arm VM (2 cores, 12 GB RAM, 200 GB disk) with Caddy for HTTPS; self-managed, idle VMs can be reclaimed, Arm capacity is often short. Fallback: a Google e2-micro VM. No server to look after: Render free (sleeps after 15 min, about 1 min wake, no disk). Small cost: Fly.io at about $2-3 a month. Fly.io has no free tier now.

Details and sources: branch `research/web-hosting`, file `docs/research/web-hosting.md`.

Decision 2026-10-08 (user): fuda keeps the web shell but picks no host. Anyone who wants it hosts it and pays. No free option fits, so $0 does not apply to the web shell.
