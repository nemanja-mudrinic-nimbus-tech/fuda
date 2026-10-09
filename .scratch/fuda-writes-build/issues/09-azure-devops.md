# 09: Azure DevOps login, Boards and writes

**What to build:** A person logs in with a Microsoft Entra ID app (confidential client on the web). fuda lists the user's organizations and their `fuda-` repositories, reads them with the user's token, and writes Move and Assign through the pushes API as compare-and-swap commits. URLs are `/azure/<org>/<project>/<repo>/`. First verify the open research items: whether users can consent to `vso.code_write` without an admin.

**Blocked by:** 05: Assign on GitHub

**Status:** in-progress: built and unit tested; the by-hand checks need a real Azure DevOps repository and Entra app

- [ ] Login, Board list, Move and Assign work against a real Azure DevOps test repo (checked by hand)
- [x] Azure source satisfies the same board interfaces as GitHub
- [x] Guide and README updated
- [x] The desktop app logs in to Azure DevOps with the device code (plain HTTP, not MSAL); the token is in the OS keychain (`FUDA_SOURCE=azure`)

Spec: `.scratch/fuda-writes/spec.md`.
