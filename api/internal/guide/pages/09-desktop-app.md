# Install the desktop app

The desktop app is a download from the latest GitHub Release. It is not signed with a developer
certificate, so your system warns the first time you open it.

## macOS

1. Download the DMG for your Mac: `arm64` for Apple Silicon, `amd64` for Intel.
2. Open it and drag fuda to Applications.
3. Open fuda from Applications. macOS says it cannot verify the app. Choose Done.
4. Open System Settings → Privacy & Security, scroll to the message about fuda and choose Open Anyway.

You only do this once. If Open Anyway does not show, run this in Terminal and open fuda again:

```sh
xattr -dr com.apple.quarantine /Applications/fuda.app
```

## Windows

Unzip the download and open `fuda-desktop.exe`. SmartScreen says "Windows protected your PC". Choose
More info, then Run anyway.

## Updates

The app checks for new releases and opens the download page when there is one. You can also choose
Check for updates in the menu.
