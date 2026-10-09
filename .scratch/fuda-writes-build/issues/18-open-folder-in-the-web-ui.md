# 18: Open folder in the web UI

**What to build:** "Open folder…" on desktop lives in the File menu today (`api/cmd/fuda-desktop/folder.go`). Move it into the web UI, so people find it where they pick a Board. The page asks the desktop shell through a small HTTP endpoint. The File menu item goes away.

**Blocked by:** None

**Status:** ready-for-agent

- [ ] The desktop app serves `POST /api/local/open-folder`. It shows the native folder dialog and adds the folder with `app.Folders.Add`. The web server does not register the endpoint
- [ ] The endpoint returns JSON: the new Board path on success, `{cancelled: true}` when the person closes the dialog, or an error code (no `tasks/` or `docs/board/tasks/` folder; any other failure with its message)
- [ ] The hosts/boards response gets a `canOpenFolder` flag. It is true only on desktop
- [ ] When `canOpenFolder` is true, the Board picker has an "Open folder…" row under the "Local" group
- [ ] When `canOpenFolder` is true and no Board is open, the empty state has an "Open folder…" button
- [ ] On success the page moves to the new Board with the router. On an error code the page shows an inline message in the app's own style. On cancel nothing happens
- [ ] On the web server the UI shows nothing: no row, no button, no hint
- [ ] The File → "Open folder…" menu item and its Cmd/Ctrl+O shortcut are removed. `folder.go` keeps only the dialog and the add-folder logic; no native warning or error pop-ups and no `window.SetURL`
- [ ] Tests: Go tests for the endpoint (success, cancel, no tasks folder, not registered without a folder-opener) and for the flag; client tests for the pure helpers
- [ ] Guide pages, README, `docs/architecture.md` and `docs/decisions.md` describe the new flow in the same change

Spec: `.scratch/fuda-writes/spec.md`.

## Decided (grilled)

- Place: the Board picker and the empty state, both.
- Bridge: HTTP endpoint, not a Wails binding. The client needs no Wails runtime.
- Web server: shows nothing. Local Boards there still come from `FUDA_LOCAL_PATH`.
- Detect: a flag in the existing hosts/boards response, not a 404 probe.
- Result: JSON plus inline page message plus router navigation, not native pop-ups.
- File menu item: removed. No replacement shortcut.
