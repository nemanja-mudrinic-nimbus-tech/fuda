# 02: GitHub login and the Board picker

**What to build:** A person logs in with the fuda GitHub App (web flow with PKCE). The token lives in an encrypted cookie. fuda lists the user's `fuda-` repositories (`GET /user/repos`) in a Board picker in the top bar and reads each Board with the user's token. `/` lists Boards, or opens the only one. The server host token and `FUDA_AUTH_USER`/`FUDA_AUTH_PASSWORD` are removed. Webhook-driven sync goes; the client polls about every 5 seconds, and the server checks the head SHA first.

**Blocked by:** 01: Many Boards with host-first URLs

**Status:** done

- [x] Not logged in: every Board page sends you to login
- [x] After login, the picker lists only `fuda-` repos the App is installed on
- [x] No access to a repository: a clear message
- [x] Expired token is refreshed; if refresh fails, back to login
- [x] Board updates within about 10 seconds of a push
- [x] README, Guide, `docs/decisions.md` and `docs/architecture.md` updated; ADR 0003 accepted

Spec: `.scratch/fuda-writes/spec.md`.
