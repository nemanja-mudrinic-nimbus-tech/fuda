package board

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"fuda/internal/taskfiles"
)

type AssignRequest struct {
	TaskID string   `json:"-"`
	Seen   []string `json:"seen"`
	Owners []string `json:"owners"`
}

func (s *Service) Assign(ctx context.Context, req AssignRequest) error {
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
	if task.Archived {
		return ErrLocked
	}

	known := newPeople(snap.develop.Config.People, snap.develop.Tasks)
	return s.commit(ctx, writer, task.Path, func(content []byte, current taskfiles.Task) ([]byte, string, error) {
		if !sameNames(known.canonical(current.Owners), known.canonical(req.Seen)) {
			who, err := writer.LastEditor(ctx, s.opts.WorkBranch, task.Path)
			if err != nil || who == "" {
				who = "Someone"
			}
			now := "no one"
			if owners := known.canonical(current.Owners); len(owners) > 0 {
				now = strings.Join(owners, ", ")
			}
			return nil, "", &ConflictError{Message: fmt.Sprintf("%s set the owners to %s just now", who, now)}
		}
		edited, err := taskfiles.ApplyAssign(content, known.canonical(req.Owners))
		return edited, assignMessage(task.ID, req.Owners), err
	})
}

func assignMessage(id string, owners []string) string {
	if len(owners) == 0 {
		return fmt.Sprintf("Remove the owners of %s", id)
	}
	return fmt.Sprintf("Assign %s to %s", id, strings.Join(owners, ", "))
}

func sameNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, name := range a {
		if !slices.Contains(b, name) {
			return false
		}
	}
	return true
}
