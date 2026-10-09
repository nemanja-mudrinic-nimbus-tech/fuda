---
status: accepted
---

# Each user reads and writes with their own host token

fuda has no server token and no shared identity. Each person logs in to GitHub (a GitHub App user token) or Azure DevOps (a Microsoft Entra ID app), and fuda reads Boards and writes commits with that token. Commits show who made them, and the host decides who may write. On the web the token is in an encrypted cookie; on desktop in the OS keychain. An account that can read but not write gets a read-only board.

## Considered Options

- One fuda bot identity that writes for everyone: rejected. Commits lose their author, and fuda must enforce access itself.

## Consequences

- Azure DevOps write access (Basic) is free for only 5 users. More writers may need paid seats. Accepted.
- GitHub cannot limit a token to one branch, and private-repo branch protection is paid. fuda itself only ever writes Task files.
- This replaces `FUDA_AUTH_USER`/`FUDA_AUTH_PASSWORD` and the server's host token.
