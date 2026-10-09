# 10: Both hosts at once

**What to build:** A person can be logged in to GitHub and Azure DevOps at once. The Board picker lists Boards from both. Login and logout are per host, with one cookie per host. A "Log out of all" button clears both. An expired token sends only that host back to login.

**Blocked by:** 09: Azure DevOps login, Boards and writes

**Status:** in-progress: built and tested; the by-hand check needs real GitHub and Azure DevOps logins

- [x] Picker shows Boards from both hosts
- [x] Logging out of one host keeps the other working
- [x] "Log out of all" clears both
- [x] The desktop app has one keychain entry per host and the same per-host logout

Spec: `.scratch/fuda-writes/spec.md`.
