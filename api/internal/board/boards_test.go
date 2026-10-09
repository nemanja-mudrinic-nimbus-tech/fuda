package board

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type fakeRepos map[string]map[string][]byte

type fakeSource struct{ files map[string][]byte }

func (f fakeSource) Head(context.Context, string) (string, error) { return "head", nil }

func (f fakeSource) Files(context.Context, string) (map[string][]byte, error) { return f.files, nil }

type missingSource struct{}

func (missingSource) Head(context.Context, string) (string, error) {
	return "", ErrBranchMissing
}

func (missingSource) Files(context.Context, string) (map[string][]byte, error) {
	return nil, ErrBranchMissing
}

func taskFile(id string) map[string][]byte {
	return map[string][]byte{"docs/board/tasks/" + id + ".md": []byte("---\nid: " + id + "\ntitle: " + id + "\nstatus: backlog\n---\n")}
}

func newTestBoards(t *testing.T, repos fakeRepos) *Boards {
	t.Helper()
	return NewBoards(slog.New(slog.DiscardHandler), func(id BoardID) (*Service, error) {
		files, ok := repos[id.Repo]
		if !ok {
			return nil, ErrNotFound
		}
		var source Source = fakeSource{files}
		if files == nil {
			source = missingSource{}
		}
		return NewService(source, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main", Origin: Origin{Path: id.Path()}}), nil
	}, nil)
}

func TestTwoBoardsKeepTheirOwnTasks(t *testing.T) {
	boards := newTestBoards(t, fakeRepos{"a/one": taskFile("ONE-1"), "a/two": taskFile("TWO-1")})

	one, err := boards.Get(context.Background(), BoardID{"github", "a/one"})
	if err != nil {
		t.Fatal(err)
	}
	two, err := boards.Get(context.Background(), BoardID{"github", "a/two"})
	if err != nil {
		t.Fatal(err)
	}

	viewOne, _ := one.Board()
	viewTwo, _ := two.Board()
	if len(viewOne.Cards) != 1 || viewOne.Cards[0].ID != "ONE-1" {
		t.Errorf("first board: %+v", viewOne.Cards)
	}
	if len(viewTwo.Cards) != 1 || viewTwo.Cards[0].ID != "TWO-1" {
		t.Errorf("second board: %+v", viewTwo.Cards)
	}
	if _, err := two.Task("ONE-1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("a task of the first board is visible in the second: %v", err)
	}
}

func TestGetReturnsTheSameServiceEachTime(t *testing.T) {
	boards := newTestBoards(t, fakeRepos{"a/one": taskFile("ONE-1")})
	first, _ := boards.Get(context.Background(), BoardID{"github", "a/one"})
	second, _ := boards.Get(context.Background(), BoardID{"github", "a/one"})
	if first != second {
		t.Error("a Board was opened twice")
	}
}

func TestGetUnknownBoardIsNotFound(t *testing.T) {
	boards := newTestBoards(t, fakeRepos{"a/gone": nil})
	for _, id := range []BoardID{
		{"github", "a/nothing"},
		{"github", "a/gone"},
		{"github", "../etc"},
		{"github", ""},
	} {
		if _, err := boards.Get(context.Background(), id); !errors.Is(err, ErrNotFound) {
			t.Errorf("%v: got %v, want not found", id, err)
		}
	}
}

type memberSource struct {
	fakeSource
	allowed string
}

func (m memberSource) Head(ctx context.Context, branch string) (string, error) {
	switch TokenFrom(ctx) {
	case "":
		return "", ErrUnauthorized
	case m.allowed:
		return m.fakeSource.Head(ctx, branch)
	}
	return "", ErrBranchMissing
}

func TestACachedBoardIsNotShownToSomeoneWithoutAccess(t *testing.T) {
	boards := NewBoards(slog.New(slog.DiscardHandler), func(BoardID) (*Service, error) {
		source := memberSource{fakeSource{taskFile("ONE-1")}, "member"}
		return NewService(source, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, nil)
	id := BoardID{"github", "a/one"}

	if _, err := boards.Get(WithToken(context.Background(), "member"), id); err != nil {
		t.Fatal(err)
	}
	if _, err := boards.Get(WithToken(context.Background(), "stranger"), id); !errors.Is(err, ErrNotFound) {
		t.Errorf("stranger: got %v, want not found", err)
	}
	if _, err := boards.Get(context.Background(), id); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("no token: got %v, want login required", err)
	}
}

type countingSource struct {
	heads atomic.Int32
	head  atomic.Value
}

func (c *countingSource) Head(context.Context, string) (string, error) {
	c.heads.Add(1)
	if h, ok := c.head.Load().(string); ok {
		return h, nil
	}
	return "one", nil
}

func (c *countingSource) Files(context.Context, string) (map[string][]byte, error) {
	return taskFile("ONE-1"), nil
}

func TestPollReadsFilesOnlyWhenTheHeadMoved(t *testing.T) {
	source := &countingSource{}
	s := NewService(source, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main", Cooldown: time.Hour})
	ctx := context.Background()

	if err := s.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	first, _ := s.Board()
	if err := s.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	second, _ := s.Board()
	if first.Sync.Develop.SyncedAt != second.Sync.Develop.SyncedAt {
		t.Error("an unchanged head caused a read")
	}

	source.head.Store("two")
	if err := s.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	third, _ := s.Board()
	if third.Sync.Develop.SHA != "two" {
		t.Errorf("a moved head was not read: %+v", third.Sync.Develop)
	}
}

func emptyBoardService(files map[string][]byte) *Service {
	return NewService(fakeSource{files}, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
}

func TestRepositoryWithoutTasksIsAnEmptyBoardWithDefaultStages(t *testing.T) {
	s := emptyBoardService(map[string][]byte{})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	view, ok := s.Board()
	if !ok {
		t.Fatal("an empty repository must still give a board")
	}
	if len(view.Cards) != 0 || len(view.Problems) != 0 {
		t.Errorf("cards %d, problems %v", len(view.Cards), view.Problems)
	}
	want := []string{"backlog", "in-progress", "in-review", "merged", "testing", "validated"}
	if got := columnIDs(view.Board); !reflect.DeepEqual(got, want) {
		t.Errorf("columns: got %v, want %v", got, want)
	}
}

func TestInvalidBoardConfigIsAProblemOnAnEmptyBoard(t *testing.T) {
	s := emptyBoardService(map[string][]byte{"docs/board/stages.md": []byte("---\nstages: [\n---\n")})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	view, _ := s.Board()
	if len(view.Problems) != 1 || view.Problems[0].Path != "docs/board/stages.md" {
		t.Errorf("problems: %v", view.Problems)
	}
	if len(view.Columns) == 0 {
		t.Error("default stages must still apply")
	}
}
