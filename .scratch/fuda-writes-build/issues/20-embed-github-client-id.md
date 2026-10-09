# 20: Embed the GitHub client id in the desktop app

**What to build:** A `fuda.app` built with `make desktop-macos` and no `GITHUB_CLIENT_ID` stops at start with "no login is set up". Finder hides that error, so the app seems dead. Make a local build start without the terminal. Tell the person when it cannot.

**Blocked by:** 19: macOS `fuda.app` bundle

**Status:** ready-for-agent

Proposal, to confirm with the user before building:

- **Embed at build time.** `make desktop-macos` and `make desktop-windows` read `FUDA_GITHUB_CLIENT_ID` (and `FUDA_AZURE_CLIENT_ID`) from the developer's local env file (the one `make dev` already loads) when `GITHUB_CLIENT_ID` is not passed, and put them in the binary with `-ldflags -X`. The client id is public, not a secret. The release workflow keeps passing it from the `FUDA_GITHUB_CLIENT_ID` repository variable.
- **Show the error.** When no login is configured, the desktop app shows a dialog with the message instead of exiting silently. The message says how to fix it.
- **No new tool.** The Makefile keeps using only `go`, `pnpm` and `docker`.

- [ ] `make desktop-macos` with the id in the local env file makes a `fuda.app` that starts from Finder (checked by hand)
- [ ] With no client id, the app shows a dialog and does not exit silently
- [ ] The release build still embeds the repository variable
- [ ] `CONTRIBUTING.md` says how a local build gets the id

Spec: `.scratch/fuda-writes/spec.md`.
