# Whose identity writes the commit, and how each shell gets the token

Type: grilling
Status: resolved
Blocked by: 03

## Question

Each Move or Assign is written as the person who made it, using their own host account. Given the research on login, which login does the web shell use and which does the desktop shell use, for GitHub and for Azure DevOps? Where is each token stored, and what does a user do the first time they open the app?

Added after research (2026-10-08): on Azure DevOps, writing needs a Basic seat, free for only 5 users. Decide whether every person writes as themselves (seats may cost money), or whether a single fuda identity writes on behalf of people without a seat, recording the real person in the commit.

## Answer

Resolved 2026-10-08 (grilling).

- Each person writes as themselves, with their own host account. On Azure DevOps, writers past the 5 free Basic seats may cost money; accepted. No shared fuda identity.
- GitHub: web shell uses the GitHub App web flow (PKCE); desktop uses device flow (no secret in the binary).
- Azure DevOps: web shell uses Entra ID as a confidential client; desktop uses MSAL device code.
- Tokens: web keeps them in an encrypted cookie (no server storage); desktop in the OS keychain.
- Login first on both shells. The web shell still reads the board with its server token; login gates the view. This replaces the shared `FUDA_AUTH_USER`/`FUDA_AUTH_PASSWORD`.
- At login fuda checks the account can see the repository; if not, it shows a message.
- An account that can see but not write gets a read-only board: cards don't drag, a note says why.
- Expired or revoked tokens: back to login (not discussed in depth; follows from the above).

## Comments

- 2026-10-08: Changed by "How a user chooses which board to look at" (10): the web shell now reads Boards with the user's own token; no server token.
