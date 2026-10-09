# 14: Desktop app icon

**What to build:** The desktop app shows the default Wails icon. Give it fuda's own icon in the window, the Dock (macOS) and the taskbar and `.exe` (Windows). The macOS `.app` bundle and its `.icns` are task 19, not here.

**Blocked by:** 04: Desktop app (Wails) first, GitHub only

**Status:** done

Decided with the user, 2026-10-09:

- **Art:** reuse `client/public/favicon.svg`. It stays the master file, so the web and desktop icons cannot drift apart.
- **Files, committed to git** (the build needs no extra tools):
  - `icon.png`: 1024 px, full square, for the window and Windows.
  - `icon-macos.png`: 1024 px, about 10% transparent margin per side (art about 824 px), keeps the favicon's rounded corners, for the Dock.
  - A `.syso` resource with the `.ico` inside, next to `main.go` in `api/cmd/fuda-desktop/` (for example `rsrc_windows_amd64.syso`). Go links it into the `.exe` on its own.
- **Run time:** embed the PNGs in the binary. Set the window icon and the Dock icon in `api/cmd/fuda-desktop/main.go`.
- **Making the files:** run once by hand. `sips` makes the PNGs from the SVG. `go run github.com/tc-hib/go-winres` makes the `.ico` and `.syso`. No Makefile target, because the Makefile uses only `go`, `pnpm` and `docker`.
- **Docs:** `CONTRIBUTING.md` lists the exact commands to remake the files. No ADR.

- [ ] `icon.png`, `icon-macos.png` and the `.syso` are committed
- [ ] The window and the macOS Dock show the fuda icon when run from `make desktop-macos` (checked by hand)
- [ ] The Windows taskbar and `fuda-desktop.exe` show the fuda icon (checked by hand)
- [ ] `CONTRIBUTING.md` has the commands to remake the icon files

Spec: `.scratch/fuda-writes/spec.md`.
