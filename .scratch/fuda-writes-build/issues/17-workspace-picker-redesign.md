# 17: Workspace picker redesign

**What to build:** Merge the origin chip (`OriginPill` in `client/src/components/organisms/AppHeader.tsx`) and the Board picker (`client/src/components/molecules/BoardPicker.tsx`) into one top-bar button. Move login and logout out of the picker into a new account menu. The UI keeps the word "Board"; "workspace" is only this task's name.

**Blocked by:** None

**Status:** done

- [x] One button replaces the chip and the arrow: host icon, Board name, arrow. The branch shows as grey text on wide screens, as the chip does today
- [x] The dropdown groups Boards under one label per host, then a "Local" group, each Board with its host icon and a check on the current one
- [x] The current Board's row has an "Open repo" action (new tab) when the Board has a URL; the chip's link is gone
- [x] With no Board open, the button reads "Select a Board". With no Boards and no hosts, there is no button
- [x] The dropdown has no login or logout items
- [x] A new account menu sits on the right of the top bar next to the theme toggle: one row per host (icon, name, "Logged in" or "Logged out", and a Log in or Log out action), plus "Log out of all" when more than one host is logged in
- [x] Guide pages and `docs/` describe the new top bar
- [x] Client tests cover the grouping, the "Select a Board" label and the "Log out of all" rule (pure helpers in `hosts.test.ts`; the client has no DOM test setup, so the rendered menus are checked by hand)

Spec: `.scratch/fuda-writes/spec.md` (story 8).

Not in this task: "Open folder…" (task 18). Leave room in the dropdown for it.
