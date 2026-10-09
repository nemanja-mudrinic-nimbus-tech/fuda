# 13: Confirm the updater's release repository

**What to build:** The desktop app looks for new versions in the GitHub repository named by `releasesRepo` in `api/cmd/fuda-desktop/update.go`. Today it is `nemanja-mudrinic-nimbus-tech/fuda`, copied from the git remote. Confirm where releases will really live. If it is another repository, change the constant, and make sure the `release` workflow runs in that repository.

**Blocked by:** 04: Desktop app (Wails) first, GitHub only

**Status:** ready-for-human

- [ ] Release repository decided
- [ ] `releasesRepo` matches it, and a tagged release is found by a running app

Spec: `.scratch/fuda-writes/spec.md`.
