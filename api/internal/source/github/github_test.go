package github

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"fuda/internal/board"
)

func serve(t *testing.T, handler http.HandlerFunc) *Source {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	s := New("o/fuda-x", "docs")
	s.base = server.URL
	s.client = server.Client()
	return s
}

func TestListBoardsKeepsOnlyFudaRepositories(t *testing.T) {
	s := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Errorf("authorization: %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`[{"full_name":"o/fuda-tasks","name":"fuda-tasks"},{"full_name":"o/api","name":"api"},{"full_name":"o/not-fuda-x","name":"not-fuda-x"}]`))
	})
	got, err := s.listBoards(board.WithToken(context.Background(), "user-token"))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"o/fuda-tasks"}) {
		t.Errorf("got %v", got)
	}
}

func TestListBoardsFollowsPages(t *testing.T) {
	pages := 0
	s := serve(t, func(w http.ResponseWriter, r *http.Request) {
		pages++
		if r.URL.Query().Get("page") == "1" {
			body := "["
			for i := range 100 {
				if i > 0 {
					body += ","
				}
				body += `{"full_name":"o/other","name":"other"}`
			}
			_, _ = w.Write([]byte(body + "]"))
			return
		}
		_, _ = w.Write([]byte(`[{"full_name":"o/fuda-late","name":"fuda-late"}]`))
	})
	got, err := s.listBoards(board.WithToken(context.Background(), "t"))
	if err != nil || !slices.Equal(got, []string{"o/fuda-late"}) || pages != 2 {
		t.Errorf("got %v, %v after %d pages", got, err, pages)
	}
}

func TestHostRefusalsBecomeBoardErrors(t *testing.T) {
	status := http.StatusUnauthorized
	s := serve(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) })
	ctx := board.WithToken(context.Background(), "t")

	if _, err := s.Head(ctx, "develop"); !errors.Is(err, board.ErrUnauthorized) {
		t.Errorf("401: %v", err)
	}
	status = http.StatusForbidden
	if _, err := s.Head(ctx, "develop"); !errors.Is(err, board.ErrForbidden) {
		t.Errorf("403: %v", err)
	}
	status = http.StatusNotFound
	if _, err := s.Head(ctx, "develop"); !errors.Is(err, board.ErrBranchMissing) {
		t.Errorf("404: %v", err)
	}
	if _, err := s.Head(context.Background(), "develop"); !errors.Is(err, board.ErrUnauthorized) {
		t.Errorf("no token: %v", err)
	}
}

func TestHeadAsksWithTheLastETag(t *testing.T) {
	calls := 0
	s := serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("If-None-Match") == `"v1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		_, _ = w.Write([]byte(`{"commit":{"sha":"abc"}}`))
	})
	ctx := board.WithToken(context.Background(), "t")
	for range 2 {
		sha, err := s.Head(ctx, "develop")
		if err != nil || sha != "abc" {
			t.Fatalf("got %q, %v", sha, err)
		}
	}
	if calls != 2 {
		t.Errorf("calls: %d", calls)
	}
}

func repositoryWithoutBranch(exists bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/o/fuda-x" && exists {
			_, _ = w.Write([]byte(`{"full_name":"o/fuda-x"}`))
			return
		}
		http.NotFound(w, r)
	}
}

func TestMissingBranchOnAReadableRepositoryIsAnEmptyBoard(t *testing.T) {
	s := serve(t, repositoryWithoutBranch(true))
	ctx := board.WithToken(context.Background(), "t")
	if _, err := s.Head(ctx, "develop"); err != nil {
		t.Fatalf("head: %v", err)
	}
	files, err := s.Files(ctx, "develop")
	if err != nil || len(files) != 0 {
		t.Fatalf("files: %v, %v", files, err)
	}
}

func TestMissingRepositoryStaysMissing(t *testing.T) {
	s := serve(t, repositoryWithoutBranch(false))
	ctx := board.WithToken(context.Background(), "t")
	if _, err := s.Head(ctx, "develop"); !errors.Is(err, board.ErrBranchMissing) {
		t.Fatalf("head: %v", err)
	}
}
