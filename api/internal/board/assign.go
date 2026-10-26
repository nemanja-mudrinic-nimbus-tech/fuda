package board

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"fuda/internal/taskfiles"
)

var ErrInvalidOwner = errors.New("an owner name must be one line of text with no commas")

type AssignRequest struct {
	TaskID     string   `json:"-"`
	FromOwners []string `json:"seen"`
	Owners     []string `json:"owners"`
}

func (s *Service) Assign(ctx context.Context, req AssignRequest) error {
	if slices.ContainsFunc(req.Owners, invalidOwner) {
		return ErrInvalidOwner
	}
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
	if task.Archived {
		return ErrNotMovable
	}

	known := newPeople(snap.develop.Config.People, snap.develop.Tasks)
	return s.commit(ctx, editor, task.Path, func(content []byte, current taskfiles.Task) ([]byte, string, error) {
		if !sameNames(known.canonical(current.Owners), known.canonical(req.FromOwners)) {
			who := s.lastEditor(ctx, editor, task.Path)
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

func invalidOwner(name string) bool {
	return strings.TrimSpace(name) == "" || strings.ContainsFunc(name, func(r rune) bool {
		return r == ',' || unicode.IsControl(r)
	})
}
