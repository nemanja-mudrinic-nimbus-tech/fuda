# How to release

A release is a GitHub Release with two macOS disk images: `fuda-desktop_darwin_arm64.dmg` (Apple
Silicon) and `fuda-desktop_darwin_amd64.dmg` (Intel). The `release` workflow
(`.github/workflows/release.yml`) builds them from a tag.

## Before the first release

Set the repository variables `FUDA_GITHUB_CLIENT_ID` and `FUDA_GITHUB_APP_SLUG`; the build fails
without them. `FUDA_GITHUB_BOARDS` and `FUDA_AZURE_BOARDS` are optional and passed to the build when set.

## Cut a release

`main` is protected, so the version bump goes through a pull request. Do every step below in that one
PR; `make check` fails if any of them is missed.

```bash
git switch -c release-0.2.1 main
echo 0.2.1 > VERSION
# also set "version" in client/package.json to 0.2.1 (see the checklist below)
git commit -am "Release 0.2.1"
git push -u origin release-0.2.1   # open the PR and merge it
git switch main && git pull
git tag v0.2.1
git push origin v0.2.1
```

### Files to bump by hand

| File | Change |
| --- | --- |
| `VERSION` | the new `X.Y.Z` (the tag is `v` plus this) |
| `client/package.json` | `"version"` equal to `VERSION` (checked by `TestClientPackageCarriesTheVersionFile`) |

Do **not** bump `BoardFormat` (`api/internal/version/version.go`) or the "Board format" mentions in
`api/internal/guide/pages`. The Board format is the version of the Board file layout, like a database
schema version: change it only when the layout changes in a way older fuda cannot read, not on every
release. A new fuda keeps reading older Boards.

- `VERSION` at the repository root is the one version. The tag must be `v` plus its contents; the
  workflow fails on a mismatch.
- `X.Y.Z` is a stable release and must be tagged on a commit that is on `main`; the workflow fails
  otherwise. `X.Y.Z-rc.N` (any `-suffix`) is a prerelease: any commit may be tagged, and the in-app
  update check does not offer it. Any other tag fails the workflow.
- The version goes into the binary (`fuda/internal/version.Fuda`) and into `Info.plist` (`CFBundleVersion`,
  `CFBundleShortVersionString`). The Makefile reads `VERSION`, so local builds carry it too;
  `VERSION=x` on the `make` line overrides it.
- The workflow first runs `make check` (the `ci` workflow), for prereleases too; nothing is created
  when it fails. Then it creates a draft release with generated notes, builds both DMGs, uploads them
  and publishes the release only after both are uploaded. If a build fails, the draft stays hidden:
  fix it and re-run the failed jobs, or delete the draft and the tag.

## What users get

The app checks the newest published release on start and from Help → Check for updates…, and when
it is newer offers to open its release page. It does not install anything. A build made with
`VERSION=dev` on the `make` line never checks; a plain local build carries the `VERSION` file and does.

The DMG is ad-hoc signed, not notarized. Installing it is described in the README.
