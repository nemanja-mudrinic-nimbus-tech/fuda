# Cost of polling for changes on GitHub and Azure DevOps

Type: research
Status: resolved
Blocked by:

## Question

fuda will check a branch head every 5 seconds per open app, up to 20 users. On GitHub, do authenticated conditional requests (ETag, 304) for a ref really not count against the rate limit, and which endpoint is cheapest? On Azure DevOps, what is the cheapest call to read a branch head, how do its rate limits (TSTUs) behave at this load, and is there an ETag or equivalent? Is there any free push signal a desktop app can receive without a public server?

## Answer

Polling every 5 seconds is affordable on both hosts. No free push signal is fit for production.

- GitHub: `GET /repos/{o}/{r}/commits/{branch}` with `Accept: application/vnd.github.sha` and `If-None-Match: "<last sha>"`. Authenticated 304s do not count against the primary limit (checked live). Avoid `git/ref/heads/{branch}`: it asks for 300 s between polls.
- Azure DevOps: `refs?filter=heads/{branch}` and pick the exact name; no ETag. About 0.01 TSTU per call against 200 per 5 minutes: under 1% of the limit.
- Poll from Go, not the webview (GitHub sends `max-age=60`). Web shell: one poll per branch on the server. Desktop: each app polls with its user's token.
- Push: webhooks need a public URL; Azure service hooks to a storage queue need a paid subscription.
- Unverified: secondary-limit cost of a 304; Azure probe was anonymous, repeat with a PAT.

Context: branch `research/change-polling-cost`, file `.scratch/fuda-writes/research/02-change-polling-cost.md`.
