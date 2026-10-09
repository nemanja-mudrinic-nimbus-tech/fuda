---
status: accepted
---

# A desktop app with Wails, next to the web server

The web shell cannot be hosted at $0 (no free host has a disk and no sleep), so fuda adds a desktop app that costs nothing to run. It is built with Wails v3 and mounts the same Go HTTP handler and React UI as the web server: one Go core, two shells. The web shell stays supported, but fuda picks no host; whoever wants it hosts it and pays.

## Consequences

- Unsigned builds warn on macOS and Windows. Avoiding the warning costs $99 a year on macOS. Windows signing is free through SignPath for open source.
- Updates come from free GitHub Releases. Wails v3 has no updater, so the app checks the latest release itself and replaces its own binary (`go-selfupdate`). Wails v3 is still beta.
- The desktop app logs in with GitHub device flow and keeps the token in the OS keychain. It has no server, no cookie and no client secret.
