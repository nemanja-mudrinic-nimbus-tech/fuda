package board

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fuda/internal/taskfiles"
)

type Writer interface {
	ReadFile(ctx context.Context, branch, path string) (content []byte, version string, err error)
	WriteFile(ctx context.Context, branch, path string, content []byte, version, message string) error
	LastEditor(ctx context.Context, branch, path string) (string, error)
}

var ErrChanged = errors.New("file changed since it was read")

var ErrLocked = errors.New("this task cannot be changed")

type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

type MoveRequest struct {
	TaskID string `json:"-"`
	Seen   string `json:"seen"`
	Column string `json:"column"`
}

const moveAttempts = 4

func (s *Service) Move(ctx context.Context, req MoveRequest) error {
	writer, err := s.requireWriter(ctx)
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
		return ErrLocked
	}
	target, ok := snap.column(req.Column)
	if !ok {
		return ErrNotFound
	}
	if target.PROpen || len(target.Statuses) == 0 {
		return ErrLocked
	}

	return s.commit(ctx, writer, task.Path, func(content []byte, current taskfiles.Task) ([]byte, string, error) {
		if normalizeStatus(current.Status) != normalizeStatus(req.Seen) {
			return nil, "", s.lostMove(ctx, writer, snap, task.Path, current.Status)
		}
		move := taskfiles.Move{Status: target.Statuses[0]}
		if current.Claimed == "" && snap.inFirstColumn(current.Status) && !snap.inFirstColumn(move.Status) {
			move.Claimed = s.now().Format(time.DateOnly)
		}
		edited, err := taskfiles.ApplyMove(content, move)
		return edited, fmt.Sprintf("Move %s to %s", task.ID, target.Name), err
	})
}

func (s *Service) commit(ctx context.Context, writer Writer, path string, edit func(content []byte, current taskfiles.Task) ([]byte, string, error)) error {
	for range moveAttempts {
		content, version, err := writer.ReadFile(ctx, s.opts.WorkBranch, path)
		if err != nil {
			return err
		}
		current, err := taskfiles.ParseOne(path, content)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		edited, message, err := edit(content, current)
		var conflict *ConflictError
		if errors.As(err, &conflict) {
			s.refresh(ctx)
			return err
		}
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		err = writer.WriteFile(ctx, s.opts.WorkBranch, path, edited, version, message)
		if errors.Is(err, ErrChanged) {
			continue
		}
		if err != nil {
			return err
		}
		s.refresh(ctx)
		return nil
	}
	s.refresh(ctx)
	return &ConflictError{Message: "This task keeps changing. Try again."}
}

func (s *Service) refresh(ctx context.Context) {
	if err := s.Sync(ctx); err != nil {
		s.opts.Logger.Warn("the board was not refreshed after a write", "error", err)
	}
}

func (s *Service) lostMove(ctx context.Context, writer Writer, snap *snapshot, path, status string) error {
	who, err := writer.LastEditor(ctx, s.opts.WorkBranch, path)
	if err != nil || who == "" {
		who = "Someone"
	}
	stage := status
	if col, ok := snap.columnOf(status); ok {
		stage = col.Name
	}
	return &ConflictError{Message: fmt.Sprintf("%s moved this to %s just now", who, stage)}
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
