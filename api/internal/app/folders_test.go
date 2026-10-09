package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"fuda/internal/board"
)

func folderWith(t *testing.T, parent, name, tasks string) string {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(tasks)), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAFolderNeedsATasksFolder(t *testing.T) {
	folders, _ := NewFolders("")
	for _, tasks := range []string{"tasks", "docs/board/tasks"} {
		if _, err := folders.Add(folderWith(t, t.TempDir(), "ok", tasks)); err != nil {
			t.Errorf("%s: %v", tasks, err)
		}
	}
	if _, err := folders.Add(t.TempDir()); !errors.Is(err, ErrNoTasks) {
		t.Errorf("an empty folder: %v", err)
	}
}

func TestTwoFoldersWithOneNameGetDifferentBoards(t *testing.T) {
	folders, _ := NewFolders("")
	first, _ := folders.Add(folderWith(t, t.TempDir(), "board", "tasks"))
	second, _ := folders.Add(folderWith(t, t.TempDir(), "board", "tasks"))

	if first == second || first.Repo != "board" || second.Repo != "board-2" {
		t.Errorf("ids %v and %v", first, second)
	}
	if _, ok := folders.Path(second); !ok {
		t.Error("the second folder is not found by its id")
	}
}

func TestAddingAFolderTwiceKeepsOneBoard(t *testing.T) {
	folders, _ := NewFolders("")
	dir := folderWith(t, t.TempDir(), "board", "tasks")
	for range 2 {
		if _, err := folders.Add(dir); err != nil {
			t.Fatal(err)
		}
	}

	if got := folders.List(); len(got) != 1 {
		t.Errorf("boards: %v", got)
	}
}

func TestFoldersAreRememberedAndMissingOnesDropped(t *testing.T) {
	store := filepath.Join(t.TempDir(), "folders.json")
	parent := t.TempDir()
	kept := folderWith(t, parent, "kept", "tasks")
	gone := folderWith(t, parent, "gone", "tasks")
	folders, _ := NewFolders(store)
	for _, dir := range []string{kept, gone} {
		if _, err := folders.Add(dir); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.RemoveAll(gone); err != nil {
		t.Fatal(err)
	}

	again, err := NewFolders(store)
	if err != nil {
		t.Fatal(err)
	}

	want := []board.BoardID{{Host: "local", Repo: "kept"}}
	if got := again.List(); len(got) != 1 || got[0] != want[0] {
		t.Errorf("remembered: %v", got)
	}
}

func TestAFolderNameBecomesAValidBoardName(t *testing.T) {
	folders, _ := NewFolders("")
	id, _ := folders.Add(folderWith(t, t.TempDir(), "_my board!", "tasks"))

	if id.Repo != "my board-" {
		t.Errorf("name: %q", id.Repo)
	}
}
