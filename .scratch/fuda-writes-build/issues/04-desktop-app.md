# 04: Desktop app (Wails) first, GitHub only

**What to build:** A Wails v3 desktop app mounts the same Go HTTP handler and React UI, so the app can be run and tried by hand while later tickets land. Login uses GitHub device flow; the token lives in the OS keychain. Board picker, Move and (once built) Assign work in the app. Releases go to GitHub Releases and the app updates itself. Azure DevOps login and "Open folder…" are not here: they join the app in the tickets that build those hosts.

**Blocked by:** 02: GitHub login and the Board picker

**Status:** ready-for-agent

- [x] App builds for macOS and Windows from `make` with only go, pnpm and docker
- [ ] Device-flow login, Board picker and Move work in the app (checked by hand)
- [x] Token is kept in the OS keychain, not on disk
- [ ] Update from a newer GitHub Release works (checked by hand)
- [x] ADR 0004 accepted; README explains the unsigned-app warnings

Later tickets add to the app: 09 adds MSAL device code and the Azure keychain entry, 10 adds per-host logout in the app, 11 adds "Open folder…" and remembered folders.

Spec: `.scratch/fuda-writes/spec.md`.
