package board

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fuda/internal/taskfiles"
)

type TaskEditor interface {
	CanWrite(ctx context.Context) (bool, error)
	ReadFile(ctx context.Context, branch, path string) (content []byte, version string, err error)
	WriteFile(ctx context.Context, branch, path string, content []byte, version, message string) error
	LastEditor(ctx context.Context, branch, path string) (string, error)
}

var ErrStale = errors.New("file changed since it was read")

var ErrNotMovable = errors.New("this task cannot be changed")

type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

type MoveRequest struct {
	TaskID     string `json:"-"`
	FromStatus string `json:"seen"`
	ToColumn   string `json:"column"`
}

const writeAttempts = 4

func (s *Service) Move(ctx context.Context, req MoveRequest) error {
	editor, err := s.requireEditor(ctx)
	if err != nil {
		return err
	}
	snap := s.snapshot.Load()
	if snap == nil {
		return ErrNotFound
	}
	task, ok := snap.byID[req.TaskID]
	if !ok {
		return ErrNotFound
	}
	if task.Archived || len(snap.reviews[task.ID]) > 0 {
		return ErrNotMovable
	}
	target, ok := snap.column(req.ToColumn)
	if !ok {
		return ErrNotFound
	}
	if target.PROpen || len(target.Statuses) == 0 {
		return ErrNotMovable
	}

	return s.commit(ctx, editor, task.Path, func(content []byte, current taskfiles.Task) ([]byte, string, error) {
		if normalizeStatus(current.Status) != normalizeStatus(req.FromStatus) {
			return nil, "", s.lostMove(ctx, editor, snap, task.Path, current.Status)
		}
		move := taskfiles.StatusEdit{Status: target.Statuses[0]}
		if current.Claimed == "" && snap.inFirstColumn(current.Status) && !snap.inFirstColumn(move.Status) {
			move.Claimed = s.now().Format(time.DateOnly)
		}
		edited, err := taskfiles.ApplyMove(content, move)
		return edited, fmt.Sprintf("Move %s to %s", task.ID, target.Name), err
	})
}

func (s *Service) commit(ctx context.Context, editor TaskEditor, path string, edit func(content []byte, current taskfiles.Task) ([]byte, string, error)) error {
	for range writeAttempts {
		content, version, err := editor.ReadFile(ctx, s.opts.WorkBranch, path)
		if err != nil {
			return err
		}
		current, err := taskfiles.ParseTask(path, content)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		edited, message, err := edit(content, current)
		var conflict *ConflictError
		if errors.As(err, &conflict) {
			s.syncAfterWrite(ctx)
			return err
		}
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if _, err := taskfiles.ParseTask(path, edited); err != nil {
			return fmt.Errorf("%s: the edit would break the file: %w", path, err)
		}
		err = editor.WriteFile(ctx, s.opts.WorkBranch, path, edited, version, message)
		if errors.Is(err, ErrStale) {
			continue
		}
		if err != nil {
			return err
		}
		s.syncAfterWrite(ctx)
		return nil
	}
	s.syncAfterWrite(ctx)
	return &ConflictError{Message: "This task keeps changing. Try again."}
}

func (s *Service) syncAfterWrite(ctx context.Context) {
	if err := s.Sync(ctx); err != nil {
		s.opts.Logger.Warn("the board was not refreshed after a write", "error", err)
	}
}

func (s *Service) lostMove(ctx context.Context, editor TaskEditor, snap *snapshot, path, status string) error {
	who := s.lastEditor(ctx, editor, path)
	stage := status
	if col, ok := snap.columnOf(status); ok {
		stage = col.Name
	}
	return &ConflictError{Message: fmt.Sprintf("%s moved this to %s just now", who, stage)}
}

func (s *Service) lastEditor(ctx context.Context, editor TaskEditor, path string) string {
	who, err := editor.LastEditor(ctx, s.opts.WorkBranch, path)
	if err != nil || who == "" {
		return "Someone"
	}
	return who
}

func (snap *snapshot) column(id string) (Column, bool) {
	for _, c := range snap.board.Columns {
		if c.ID == id {
			return c, true
		}
	}
	return Column{}, false
}

func (snap *snapshot) columnOf(status string) (Column, bool) {
	for _, c := range snap.board.Columns {
		for _, s := range c.Statuses {
			if normalizeStatus(s) == normalizeStatus(status) {
				return c, true
			}
		}
	}
	return Column{}, false
}

func (snap *snapshot) inFirstColumn(status string) bool {
	col, ok := snap.columnOf(status)
	return ok && len(snap.board.Columns) > 0 && col.ID == snap.board.Columns[0].ID
}
