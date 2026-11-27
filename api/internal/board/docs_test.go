package board

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDocsListsMarkdownOutsideTasksAndArchive(t *testing.T) {
	s := emptyBoardService(map[string][]byte{
		"docs/architecture.md":             []byte("# A"),
		"docs/board/README.md":             []byte("# B"),
		"docs/board/stages.md":             []byte("---\nstages: [backlog, done]\n---\n"),
		"docs/board/tasks/one.md":          []byte(taskContent("backlog", "")),
		"docs/board/tasks/epic/two.md":     []byte(taskContent("backlog", "")),
		"docs/board/tasks/epic/map.md":     []byte("# Map"),
		"docs/board/archive/old.md":        []byte(taskContent("done", "")),
		"docs/board/archive/epic/older.md": []byte(taskContent("done", "")),
		"docs/guides/setup.md":             []byte("# S"),
		"docs/guides/diagram.png":          []byte("png"),
		"src/README.md":                    []byte("# not under docs"),
	})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"architecture.md", "board/README.md", "guides/setup.md"}
	if got := s.Docs(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDocIsNotFoundForABoardConfigFile(t *testing.T) {
	s := emptyBoardService(map[string][]byte{
		"docs/board/stages.md": []byte("---\nstages: [backlog, done]\n---\n"),
		"docs/board/TASKS.md":  []byte("# Rules"),
	})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Doc("board/stages.md"); !errors.Is(err, ErrNotFound) {
		t.Errorf("config file: got %v, want ErrNotFound", err)
	}
	if _, err := s.Doc("board/TASKS.md"); err != nil {
		t.Errorf("TASKS.md: %v", err)
	}
}

func TestDocsIsEmptyBeforeTheFirstSync(t *testing.T) {
	s := emptyBoardService(map[string][]byte{})
	if got := s.Docs(); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}
