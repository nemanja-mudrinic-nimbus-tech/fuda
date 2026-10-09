# How the board shows a Move that is not yet confirmed

Type: grilling
Status: resolved
Blocked by: 

## Question

A user drops a card. The commit takes time and can fail (network down, token expired, no write access). Does the card move at once (optimistic) or wait? What does the user see while it waits, on success, and on failure? How does this sit with the conflict rule in "Two changes to the same Task at once" (card snaps back with a message)?

## Answer

Resolved 2026-10-08 (grilling).

- Optimistic: the card moves at once on drop. It goes back if the write fails.
- While saving: a small "saving" dot on the card. The card cannot be dragged again until the save ends.
- On failure: the card goes back with a short message saying why. An expired login shows a "log in again" link. This matches the conflict rule in "Two changes to the same Task at once".
- Leaving the page while a save is in progress: the browser warns first.
