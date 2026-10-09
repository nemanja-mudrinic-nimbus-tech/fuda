# 16: Warn when people are not defined

**What to build:** When a Board has no `people.md`, or it lists no people, and no Task has an Owner yet, the Owner picker shows "No one found" and the user does not know why. Show a clear hint in the picker instead: "No people yet. Add `docs/board/people.md` (see the Guide)", with a link to the Guide page that explains the file. When people come only from names on the board (no `people.md`), the picker may show a short note that `people.md` gives a fixed list.

**Blocked by:** 05: Assign on GitHub

**Status:** needs-info

- [ ] The picker shows the hint, not "No one found", when the list of people is empty
- [ ] The Guide page names the file, its path and a small example
- [ ] A client test covers the choice between the hint and the list

Spec: `.scratch/fuda-writes/spec.md`.

## Open questions (to grill)

The hint should also link to a modal (or similar) that explains how to fix missing people. Not decided yet: where the warning shows, what the modal holds, who sees it, and whether the Guide link stays.
