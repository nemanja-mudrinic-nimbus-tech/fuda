# 23: Sharp desktop icons

**What to build:** The desktop app's PNG icons look blurry. Find out why and make them sharp in the Dock, the window, and the Windows taskbar. Cause is not known yet.

**Blocked by:** None

**Status:** needs-triage

Where the icons are (all in `api/cmd/fuda-desktop/`):

- `icon.png` and `icon-macos.png`, both 1024x1024 RGBA. `appIcon()` in `main.go` picks one at runtime.
- `icon.icns` is copied into `fuda.app` by the Makefile (line 51).
- `rsrc_windows_amd64.syso` and `rsrc_windows_arm64.syso` hold the Windows icon.
- The web app uses `client/public/favicon.svg`, which is a vector and not part of this task.

Ideas to check first:

- The PNGs may be scaled up from a small source. Look at the original pixels at 100%.
- `icon.icns` and the `.syso` files may lack the small sizes (16, 32, 64, 128, 256, 512, plus @2x). The OS then scales one size down and it looks soft.
- The macOS icon may need the standard padding and rounded shape, or it is drawn at the wrong scale.
- The runtime icon passed to Wails may be shown at a small size from a very large image without good downscaling.

- [ ] Cause found and written in the PR description
- [ ] A vector or high-resolution source file is kept in the repo (for example `api/cmd/fuda-desktop/icon-source.svg`) with the command that makes every size
- [ ] `icon.icns` has all standard sizes; the Windows resource has 16, 24, 32, 48, 64 and 256
- [ ] The Dock, the app switcher, the window and the Windows taskbar show a sharp icon (screenshots from a real build)
- [ ] `make check` passes; `CONTRIBUTING.md` says how to regenerate the icons

Origin: seen while testing the desktop app.
