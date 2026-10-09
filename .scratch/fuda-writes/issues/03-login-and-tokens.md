# User login and tokens for GitHub and Azure DevOps

Type: research
Status: resolved
Blocked by:

## Question

How does each user log in so fuda can write as them? For GitHub: OAuth App versus GitHub App user tokens, the device flow for a desktop app, the web flow for a server, token lifetime and refresh, and the smallest scopes that allow writing to one branch. For Azure DevOps: Microsoft Entra ID sign-in for a desktop app and for a web server, versus a personal access token, and the scopes needed. What must be registered, by whom, and does any of it cost money?

## Answer

Login is free to set up on both hosts. Azure DevOps seats are the real cost.

- GitHub: use a GitHub App with user tokens (Contents: write only on installed repos; commits are the user's). Access token 8 h, refresh 6 months. Server: web flow with PKCE. Desktop: code + PKCE on a loopback redirect (secret ships in the binary) or device flow (no secret).
- No GitHub token can be limited to one branch, and branch protection on private repos is paid. fuda itself must keep writes on the board branch.
- Azure DevOps: old Azure DevOps OAuth is closed to new apps and ends in 2026. Use a free Microsoft Entra ID app registration with `vso.code_write`. Server: confidential client. Desktop: MSAL Go, loopback or device code. Refresh 90 days. Personal Microsoft accounts need a PAT instead.
- Cost: GitHub App, Entra registration and Entra ID Free are $0. **Writing code on Azure DevOps needs Basic access, free for only 5 users.** Up to 15 more writers could need paid seats.
- Unverified: loopback ports for GitHub Apps; whether users can consent to `vso.code_write` without an admin; Basic seat price.

Context: branch `research/login-and-tokens`, file `.scratch/fuda-writes/research/03-login-and-tokens.md`.
