# 06: Read-only and archived cards

**What to build:** An account that can read but not write a repository gets a read-only Board: cards don't drag and a note says why. Archived Tasks never drag, and nothing can be dropped into or out of the archive.

**Blocked by:** 03: Move on GitHub

**Status:** done

- [x] Write access is checked when a Board opens
- [x] Board service refuses Move and Assign for read-only accounts and archived Tasks (tests)
- [x] UI shows the read-only note and no drag handles

Spec: `.scratch/fuda-writes/spec.md`.
