# 26: macOS says `fuda.app` is damaged

**What to build:** A `fuda.app` sent to another Mac must open after the usual "unverified developer" prompt. Today macOS says it "is damaged and can't be opened" and offers no way past it. Ship the macOS app as a DMG that holds a validly ad-hoc-signed `fuda.app`. Stay without an Apple Developer ID.

**Blocked by:** none

**Status:** ready-for-human

Cause, checked on `bin/fuda.app` on Apple Silicon, 2026-10-22: the Go linker ad-hoc signs only the Mach-O (`flags=adhoc,linker-signed`, `Info.plist=not bound`, `Sealed Resources=none`). `make desktop-macos` then adds `Info.plist` and `Resources/` around it, so the bundle no longer matches its signature. `codesign --verify --deep --strict` fails with "code has no resources but signature indicates they must be present". A locally built copy has no quarantine flag, so Gatekeeper never checks it. A copy sent by AirDrop, chat or the GitHub release is quarantined, Gatekeeper finds the invalid signature, and blocks it as damaged. An unsigned or validly ad-hoc-signed app only gets the prompt. Running `codesign --force --sign - --identifier io.nimbus-tech.fuda` on a copy of the bundle made the check pass.

Decided with the user, 2026-10-22:

- **Signing:** the last step of `desktop-macos` runs `codesign --force --sign - --identifier io.nimbus-tech.fuda bin/fuda.app`, after `Info.plist` and the icon are in place. Nothing may change the bundle after this step.
- **DMG:** a new `desktop-macos-dmg` target that depends on `desktop-macos`. It copies `fuda.app` and a symlink to `/Applications` into a staging folder and runs `hdiutil create -format UDZO -srcfolder`, giving `bin/fuda-desktop_darwin_<arch>.dmg`. Plain Finder layout, no background. The DMG itself is not signed.
- **Release:** `.github/workflows/release.yml` builds `desktop-macos-dmg` and uploads the DMG instead of the `.tar.gz`, for arm64 and amd64.
- **Update check:** `api/cmd/fuda-desktop/update.go` only opens the release page, but it finds the release with `selfupdate.DetectLatest`, which counts a release only if it has an asset for this OS and arch in a format go-selfupdate knows (`.zip`, `.tar.gz`, …, not `.dmg`). With a DMG-only release, macOS users would never hear of a new version. Decided: find the newest release and compare its version directly, with no matching asset needed, and keep opening the release page. The release ships only the DMG for macOS.
- **Makefile rule:** the macOS targets also use `codesign` and `hdiutil`, which ship with macOS and the Xcode Command Line Tools (needed for cgo anyway). Add that exception to the rule in `CLAUDE.md` and `CONTRIBUTING.md`.
- **Docs:** README install section and a Guide page say how to open the app the first time. Open the DMG and drag fuda to Applications. On macOS 15+, open it once, then System Settings → Privacy & Security → Open Anyway. As a fallback, `xattr -dr com.apple.quarantine /Applications/fuda.app`. This replaces the "right-click, Open" line from ticket 19, which macOS 15 removed.
- **Out of scope:** Developer ID signing and notarization (stays TBD in `docs/decisions.md`), a universal binary, a styled DMG, Windows.

- [x] `make desktop-macos` produces a `fuda.app` that passes `codesign --verify --deep --strict`
- [x] `make desktop-macos-dmg` produces a DMG with `fuda.app` and an Applications link
- [x] The release uploads the DMG for arm64 and amd64, and no macOS `.tar.gz`
- [x] The update check finds a new release on macOS
- [x] `CLAUDE.md`, `CONTRIBUTING.md`, README, the Guide page and `docs/decisions.md` match the new behaviour
- [ ] A DMG copied to a second Apple Silicon Mac (or AirDropped to yourself) shows Open Anyway, not "damaged" (checked by hand)

Follows 19: macOS `fuda.app` bundle. Research: `.scratch/fuda-writes/issues/18-desktop-distribution.md`.
