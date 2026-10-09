# How board-branch changes reach develop

Type: grilling
Status: resolved
Blocked by: 05

## Question

Moves and Assigns land on the board branch. When and how do they reach develop, so the files agents and developers read are current: on a timer, on demand, through a PR, by a direct commit, by a GitHub Action or Azure pipeline, or by the app? Whose identity makes that commit?

## Answer

No longer needed. "Which branch holds a Task's current Status" moved Tasks to their own repository, so there is no board branch to merge back.
