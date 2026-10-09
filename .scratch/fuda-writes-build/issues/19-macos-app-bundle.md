# 19: macOS `fuda.app` bundle

**What to build:** The macOS release is a bare binary `fuda-desktop` in a `.tar.gz`. Ship a real `fuda.app` instead, so Finder and the Dock show the fuda icon.

**Blocked by:** 14: Desktop app icon

**Status:** ready-for-agent

Decided with the user, 2026-10-09:

- **Bundle:** `fuda.app` with an `Info.plist` and a `.icns`. Make the `.icns` from the icon PNGs with `iconutil` and commit it. Build the bundle from a `make` target, with no tool beyond `go`, `pnpm` and `docker`.
- **Release:** `.github/workflows/release.yml` packs a `.tar.gz` that holds `fuda.app`.
- **Self-update:** keep `api/cmd/fuda-desktop/update.go` as it is. It replaces the inner binary at `fuda.app/Contents/MacOS/fuda-desktop`. This was never tested inside a bundle, so check it by hand once on a Mac.
- **Signing:** stay unsigned. Update the README: "right-click, Open" the first time. Add real signing and notarization (needs an Apple Developer account) under TBD in `docs/decisions.md`, with a proposal.
- **Windows:** unchanged. The `.exe` stays a bare file.

- [x] `make desktop-macos` produces `fuda.app` that shows the icon in Finder and the Dock
- [x] The release archive holds `fuda.app`
- [ ] Update from a newer GitHub Release works with the bundle (checked by hand)
- [x] README and `docs/decisions.md` match the new behaviour

Spec: `.scratch/fuda-writes/spec.md`.
