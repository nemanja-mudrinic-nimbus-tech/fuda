# Writing to a board branch through the GitHub and Azure DevOps APIs

Type: research
Status: resolved
Blocked by:

## Question

How does fuda make one commit that changes one or more Task files on a branch, through the GitHub REST API and the Azure DevOps REST API, without a local git checkout? For each host: the endpoints and the number of calls per commit, how a stale parent is detected (compare-and-swap on the branch head), write rate limits including GitHub's secondary limits on content creation, whether a non-default branch or a custom ref (`refs/fuda/*`) works, and what permissions the user's token needs. Is about one commit every few seconds, from up to 20 users, within the limits?

## Answer

One commit per change is easy and well within limits on both hosts, as long as each user writes with their own token.

- GitHub: 1 GraphQL call (`createCommitOnBranch` with `expectedHeadOid`: true compare-and-swap, branches only) or 3 REST writes (trees, commits, refs with `force:false`: fast-forward check only).
- Azure DevOps: 1 call, `POST .../pushes` with `refUpdates[].oldObjectId` (true compare-and-swap) and N `edit` changes.
- A normal non-default branch works on both. A custom ref (`refs/fuda/*`) is unproven: likely on GitHub REST, not GraphQL, undocumented on Azure.
- Permissions: GitHub Contents: write (fine-grained) or `repo`; Azure `vso.code_write` plus Contribute. Azure global PATs stop working on 1 December 2026.
- Limits: GitHub content creation 80/min and 500/hr per user; Azure 200 TSTUs per 5 min per user. One shared token would break GitHub's hourly cap.
- Main risk: compare-and-swap conflicts on the shared branch. Serialize writes per branch, retry by re-reading the head, batch a user's rapid edits.
- 7 items unverified; listed in the research file.

Context: branch `research/board-branch-writes`, file `.scratch/fuda-writes/research/01-board-branch-writes.md`.
