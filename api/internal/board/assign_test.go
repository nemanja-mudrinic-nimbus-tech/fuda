package board

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestAssignWritesTheOwners(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "labels: a\n")})
	s := newMoveService(t, host)

	err := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Seen: nil, Owners: []string{"Ann", "Ben"}})
	if err != nil {
		t.Fatal(err)
	}

	check(t, "file", host.file(), taskContent("backlog", "owner: Ann, Ben\nlabels: a\n"))
	check(t, "writes", host.writes, 1)
	view, _ := s.Board()
	check(t, "owners on the card", card(view.Board, "T-1").Owners, []string{"Ann", "Ben"})
}

func TestAssignRemovingEveryOwnerDeletesTheLine(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "owner: Ann\n")})
	s := newMoveService(t, host)

	if err := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Seen: []string{"Ann"}}); err != nil {
		t.Fatal(err)
	}

	check(t, "file", host.file(), taskContent("backlog", ""))
}

func TestAssignRetriesSilentlyWhenAnotherLineChanged(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "owner: Ann\nlabels: a\n")})
	s := newMoveService(t, host)
	var once sync.Once
	host.beforeWrite = func(h *fakeHost) {
		once.Do(func() {
			h.edit(func(c string) string { return strings.Replace(c, "labels: a", "labels: a, b", 1) })
		})
	}

	if err := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Seen: []string{"Ann"}, Owners: []string{"Ben"}}); err != nil {
		t.Fatal(err)
	}

	check(t, "file", host.file(), taskContent("backlog", "owner: Ben\nlabels: a, b\n"))
}

func TestAssignLosesWhenOwnersChangedFirst(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "owner: Ann\n")})
	host.editor = "Ben"
	s := newMoveService(t, host)
	var once sync.Once
	host.beforeWrite = func(h *fakeHost) {
		once.Do(func() {
			h.edit(func(c string) string { return strings.Replace(c, "owner: Ann", "owner: Cy", 1) })
		})
	}

	err := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Seen: []string{"Ann"}, Owners: []string{"Di"}})

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("got %v, want a conflict", err)
	}
	check(t, "message", conflict.Message, "Ben set the owners to Cy just now")
	check(t, "file", host.file(), taskContent("backlog", "owner: Cy\n"))
	check(t, "writes", host.writes, 0)
}

func TestAssignSeesTheSameOwnersWhateverTheOrderOrAlias(t *testing.T) {
	host := newFakeHost(map[string]string{
		taskPath:               taskContent("backlog", "owner: Ben, ann\n"),
		"docs/board/people.md": "---\npeople:\n  - name: Ann\n    aliases: [ann]\n  - name: Ben\n---\n",
	})
	s := newMoveService(t, host)

	if err := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Seen: []string{"Ann", "Ben"}, Owners: []string{"Ben"}}); err != nil {
		t.Fatal(err)
	}

	check(t, "file", host.file(), taskContent("backlog", "owner: Ben\n"))
}

func TestAssignRefusals(t *testing.T) {
	archivePath := "docs/board/archive/T-2.md"
	host := newFakeHost(map[string]string{
		taskPath:    taskContent("backlog", ""),
		archivePath: "---\nid: T-2\ntitle: Two\nstatus: done\n---\n",
	})
	s := newMoveService(t, host)

	tests := []struct {
		name string
		id   string
		want error
	}{
		{"an archived task", "T-2", ErrLocked},
		{"an unknown task", "T-9", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Assign(context.Background(), AssignRequest{TaskID: tt.id, Owners: []string{"Ann"}}); !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
	check(t, "writes", host.writes, 0)
}

func TestAssignNeedsASourceThatWrites(t *testing.T) {
	s := NewService(fakeSource{files: map[string][]byte{taskPath: []byte(taskContent("backlog", ""))}}, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
	if err := s.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	err := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Owners: []string{"Ann"}})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("got %v, want ErrForbidden", err)
	}
}
