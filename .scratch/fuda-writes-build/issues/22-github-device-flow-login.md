# 22: One GitHub device-flow login for web and desktop

**What to build:** GitHub login works the same on web and desktop: device flow, client id only. The web server drops the GitHub redirect flow, `FUDA_GITHUB_CLIENT_SECRET`, and (for GitHub) `FUDA_COOKIE_SECRET` and `FUDA_BASE_URL`. Azure keeps its current login.

**Blocked by:** none

**Status:** ready-for-agent

Decided with the user:

- **One flow.** Device flow for GitHub on web and desktop. The GitHub App has "Enable Device Flow" on and "Expire user authorization tokens" off, so there is no refresh code and no secret.
- **Desktop.** Token stays in the OS keychain (`api/internal/keychain`).
- **Web.** Token stays in a sealed HttpOnly cookie. The server makes its own key on first start and saves it in `FUDA_CACHE_DIR` with mode 0600. No `FUDA_COOKIE_SECRET` for GitHub.
- **Secure cookie.** Always `Secure`, except on localhost. A plain-HTTP LAN address gets no working login; docs say so.
- **UI.** The Sign in button in the app bar starts login. No automatic login screen.
- **Old code.** Delete the GitHub redirect flow. Ignore the removed vars and log one warning. `FUDA_COOKIE_SECRET` and `FUDA_BASE_URL` stay for Azure only.
- **Docker.** Declare `FUDA_CACHE_DIR` as a volume, or every restart logs users out. Say so in the self-host Guide page.

- [ ] Web: with only `FUDA_GITHUB_CLIENT_ID` set, Sign in works and survives a server restart
- [ ] Desktop: Sign in works and the token is in the keychain
- [ ] Web and desktop show the same device-code screen
- [ ] Removed vars are ignored with one warning; Azure login still works
- [ ] Dockerfile declares the cache volume
- [ ] `make check` passes
- [ ] README, Guide self-host page, `.env.example`, `CONTRIBUTING.md`, `docs/architecture.md` and `docs/decisions.md` match the code (move the TBD item "One device-flow login for GitHub" to Settled)
