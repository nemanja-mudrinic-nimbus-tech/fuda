# Glossary

The words fuda uses, and where each one lives in the code.

## Language

**Board**:
One tasks repository, shown as Stages and Tasks. One fuda can show several Boards.
In code: `board.Service` is one open Board, `board.Registry` holds them, `board.ID` names one.

**Local Board**:
A Board read from a folder on disk instead of from a git host. A Move or an Assign changes the file on disk; the person commits it. Only the desktop app writes to a Local Board.
In code: host `board.HostLocal`, `source/local`, folders kept by `workspace.Folders`.

**Task**:
One unit of work, stored as one markdown file in the repository.
In code: `taskfiles.Task` as parsed, `board.Card` as drawn.
_Avoid_: ticket, issue, card (a card is only how a Task is drawn on the board)

**Status**:
The field on a Task that says where it is in the workflow.
In code: the `status` line, changed by `taskfiles.StatusEdit`.

**Stage**:
A column on the board. One Stage groups one or more Statuses.
In code: `board.Column`, configured in `stages.md`.
_Avoid_: lane, state

**Owner**:
The person or people working on a Task.
In code: the `owner` line, changed by `taskfiles.ApplyAssign`.
_Avoid_: assignee

**Move**:
Changing a Task's Status from fuda, so it appears in another Stage.
In code: `board.Service.Move`.

**Assign**:
Changing a Task's Owner from fuda.
In code: `board.Service.Assign`.

**Problem**:
A repository file fuda could not read. It is shown to people, never fatal.
In code: `board.Problem`.

**Activity entry**:
One line of a Task's `## Activity` section: time, actor, event, result and an optional usage. A line in another form keeps only its text. Lines indented under it (two spaces, then `- `) are its **detail lines**. The task panel shows all entries as a timeline with their detail lines; a Stage with `activity: true` shows the latest on its cards, without detail lines.
In code: `taskfiles.Activity` (`Details` for detail lines), parsed by `taskfiles.SplitActivity`; `board.Card.Activity`, `board.TaskView.Activities`; client `ActivityTimeline`, `lib/activity.ts`.

**In review**:
The Stage a Task is in while an open pull request names it. The pull request, not the Status, puts it there.
In code: `board/reviews.go`, `board.LinkedPR`.

**Read-only Board**:
A Board the logged-in person cannot write to. It shows no drag handles and refuses Move and Assign.
In code: `Service.CanWrite`, `TaskEditor.CanWrite`.

**Host**:
Where a Board lives: `github`, `azure` or `local`.
In code: `board.Host`, `FUDA_HOSTS`.

**Board id**:
A host and a repository, for example `github/owner/repo`. It is also the path of the Board's URL.
In code: `board.ID`.

**Login**:
How a person proves who they are to a host: device flow (GitHub) or OAuth redirect with PKCE (Azure). The token is kept in a sealed cookie on the web and in the OS keychain on desktop.
In code: `login.NewCookieDevice`, `login.NewCookieOAuth`, `login.NewKeychainDevice`.

**Token**:
The host's access token for one person. fuda has no server token; every read and write uses the caller's own.
In code: `board.WithToken`, `board.TokenFrom`.

## Relationships

- A **Stage** groups one or more **Statuses**; a **Task** has exactly one **Status**.
- A **Move** and an **Assign** are the only changes fuda makes to a **Task**. Creating and editing Task text happens in git.

## Flagged ambiguities

- "ticket" and "issue" were used for **Task**. Resolved: the term is **Task**.
- "assign" was used for setting the **Owner**. fuda's file field is `owner`; the action is **Assign**.
