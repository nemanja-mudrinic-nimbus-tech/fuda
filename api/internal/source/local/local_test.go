package local_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fuda/internal/board"
	"fuda/internal/source/local"
)

func taskFile(extra string) string {
	return "---\nid: T-1\ntitle: One\nstatus: backlog\n" + extra + "---\n\nBody\n"
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func openBoard(t *testing.T, root string) *board.Service {
	t.Helper()
	s := board.NewService(local.New(root, "docs", "docs/board", "develop"), board.Options{
		DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main",
	})
	if err := s.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func read(t *testing.T, root, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func column(t *testing.T, s *board.Service, id string) string {
	t.Helper()
	view, _ := s.Board()
	for _, c := range view.Cards {
		if c.ID == id {
			return c.Column
		}
	}
	t.Fatalf("no card %s", id)
	return ""
}

func TestBothLayoutsMoveAndAssignOnDisk(t *testing.T) {
	for name, path := range map[string]string{
		"code repository":  "docs/board/tasks/T-1.md",
		"tasks repository": "tasks/T-1.md",
	} {
		t.Run(name, func(t *testing.T) {
			root := writeTree(t, map[string]string{path: taskFile("labels: a\n")})
			s := openBoard(t, root)

			if err := s.Move(context.Background(), board.MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"}); err != nil {
				t.Fatal(err)
			}
			if err := s.Assign(context.Background(), board.AssignRequest{TaskID: "T-1", Owners: []string{"Ann"}}); err != nil {
				t.Fatal(err)
			}

			got := read(t, root, path)
			for _, want := range []string{"status: in progress\n", "owner: Ann\n", "labels: a\n"} {
				if !strings.Contains(got, want) {
					t.Errorf("file lacks %q:\n%s", want, got)
				}
			}
			if got := column(t, s, "T-1"); got != "in-progress" {
				t.Errorf("column after the move: %q", got)
			}
		})
	}
}

func TestAnEditorChangeToAnotherLineIsKept(t *testing.T) {
	root := writeTree(t, map[string]string{"tasks/T-1.md": taskFile("")})
	s := openBoard(t, root)
	if err := os.WriteFile(filepath.Join(root, "tasks/T-1.md"), []byte(taskFile("labels: edited\n")), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.Move(context.Background(), board.MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"}); err != nil {
		t.Fatal(err)
	}

	got := read(t, root, "tasks/T-1.md")
	if !strings.Contains(got, "labels: edited\n") || !strings.Contains(got, "status: in progress\n") {
		t.Errorf("file:\n%s", got)
	}
}

func TestAnEditorChangeToTheSameFieldWinsAndTheMoveIsRefused(t *testing.T) {
	root := writeTree(t, map[string]string{"tasks/T-1.md": taskFile("")})
	s := openBoard(t, root)
	edited := "---\nid: T-1\ntitle: One\nstatus: testing\n---\n\nBody\n"
	if err := os.WriteFile(filepath.Join(root, "tasks/T-1.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	err := s.Move(context.Background(), board.MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})

	var conflict *board.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("want a conflict, got %v", err)
	}
	if got := read(t, root, "tasks/T-1.md"); got != edited {
		t.Errorf("the editor's file was overwritten:\n%s", got)
	}
}

func TestEditorChangesShowOnThePollAfterThem(t *testing.T) {
	root := writeTree(t, map[string]string{"tasks/T-1.md": taskFile("")})
	s := openBoard(t, root)
	if err := os.WriteFile(filepath.Join(root, "tasks/T-1.md"), []byte("---\nid: T-1\ntitle: One\nstatus: testing\n---\n\nBody\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := column(t, s, "T-1"); got != "testing" {
		t.Errorf("the card is in %q", got)
	}
}

func TestALocalBoardHasNoInReview(t *testing.T) {
	root := writeTree(t, map[string]string{"tasks/T-1.md": taskFile("pr: 7\n"), "tasks/repos.md": "---\nrepos:\n  - acme/code\n---\n"})
	s := openBoard(t, root)

	view, _ := s.Board()
	for _, c := range view.Cards {
		if len(c.OpenPRs) > 0 {
			t.Errorf("card %s has open PRs: %v", c.ID, c.OpenPRs)
		}
	}
}

func TestAFolderThatCannotBeWrittenIsReadOnly(t *testing.T) {
	root := writeTree(t, map[string]string{"tasks/T-1.md": taskFile("")})
	s := openBoard(t, root)
	if err := os.Chmod(filepath.Join(root, "tasks"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "tasks"), 0o755) })
	if f, err := os.CreateTemp(filepath.Join(root, "tasks"), "probe"); err == nil {
		_ = f.Close()
		t.Skip("this user can write to read-only folders")
	}

	err := s.Move(context.Background(), board.MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})

	if !errors.Is(err, board.ErrForbidden) {
		t.Errorf("want forbidden, got %v", err)
	}
}
