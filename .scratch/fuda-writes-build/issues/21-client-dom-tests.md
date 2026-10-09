# 21: Client DOM tests

**What to build:** The client tests run in plain Node (`environment: 'node'` in `client/vite.config.ts`), so they only cover pure functions. Add a DOM test setup so components can be rendered and tested. Then add the component tests that task 17 could not write.

**Blocked by:** None

**Status:** ready-for-agent

- [ ] A DOM environment (for example `jsdom`) and `@testing-library/react` are dev dependencies. Pure-function tests keep running fast; use a per-file `// @vitest-environment jsdom` or a `projects` split, not a global switch if that slows them
- [ ] A small test helper renders a component with a `QueryClient` seeded with a Board listing
- [ ] `WorkspacePicker` tests: no button with no Boards and no hosts; "Select a Board" with no Board open; Boards grouped under a label per host then "Local"; a check on the current Board; "Open repo" only when the Board has a URL, with `target="_blank"`
- [ ] `AccountMenu` tests: no button with no hosts; one row per host showing "Logged in" or "Logged out" with a Log in or Log out action; "Log out of all" only when more than one host is logged in
- [ ] `make check` still passes; `CONTRIBUTING.md` says how to write a component test

Origin: task 17 left these out because the client had no DOM setup.
