# A user logged in to GitHub and Azure DevOps at once

Type: grilling
Status: resolved
Blocked by: 10

## Question

With many Boards found from the user's login, can one user be logged in to both hosts at once and see Boards from both in one picker? How do URLs keep them apart (`/<owner>/<repo>/` may clash), and how does login and logout work per host?

## Answer

Resolved 2026-10-08 (grilling).

- One person can be logged in to GitHub and Azure DevOps at once. The Board picker lists Boards from both hosts together.
- The host comes first in the URL: `/github/<owner>/<repo>/...` and `/azure/<org>/<project>/<repo>/...`. A Local Board follows the same idea (`/local/<folder>/...`). This replaces the `/<owner>/<repo>/` URL from "How a user chooses which board to look at".
- Login and logout are per host. Each host token is stored and expires on its own (web: one encrypted cookie per host; desktop: one keychain entry per host). A "Log out of all" button clears every host.
- An expired token sends only that host back to login; Boards on the other host keep working.
