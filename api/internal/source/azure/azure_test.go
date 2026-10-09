package azure

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"fuda/internal/board"
)

type fakeHost struct {
	head     string
	blob     string
	content  string
	pushes   []map[string]any
	moveHead bool
	refuse   bool
}

func (f *fakeHost) handle(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Errorf("authorization: %q", r.Header.Get("Authorization"))
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/refs"):
			_, _ = w.Write([]byte(`{"value":[{"name":"refs/heads/develop","objectId":"` + f.head + `"}]}`))
		case strings.HasSuffix(r.URL.Path, "/items"):
			_, _ = w.Write([]byte(`{"objectId":"` + f.blob + `","content":` + quote(f.content) + `}`))
		case strings.HasSuffix(r.URL.Path, "/pushes"):
			var push map[string]any
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &push)
			if f.refuse {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if f.moveHead {
				f.moveHead = false
				f.head = "head-2"
				w.WriteHeader(http.StatusConflict)
				return
			}
			f.pushes = append(f.pushes, push)
			w.WriteHeader(http.StatusCreated)
		case strings.HasSuffix(r.URL.Path, "/commits"):
			_, _ = w.Write([]byte(`{"value":[{"author":{"name":"Ben"}}]}`))
		case strings.HasSuffix(r.URL.Path, "/profiles/me"):
			_, _ = w.Write([]byte(`{"id":"me"}`))
		case strings.HasSuffix(r.URL.Path, "/_apis/accounts"):
			_, _ = w.Write([]byte(`{"value":[{"accountName":"acme"}]}`))
		case r.URL.Path == "/acme/_apis/git/repositories":
			_, _ = w.Write([]byte(`{"value":[{"name":"fuda-tasks","project":{"name":"Team"}},{"name":"api","project":{"name":"Team"}}]}`))
		case strings.Contains(r.URL.Path, "/_apis/permissions/"):
			_, _ = w.Write([]byte(`{"value":[true]}`))
		case strings.HasSuffix(r.URL.Path, "/fuda-tasks"):
			_, _ = w.Write([]byte(`{"id":"repo-id","project":{"id":"project-id"}}`))
		default:
			t.Errorf("unexpected request %s", r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func serve(t *testing.T, f *fakeHost) *Source {
	t.Helper()
	server := httptest.NewServer(f.handle(t))
	t.Cleanup(server.Close)
	s := New(Repo{Org: "acme", Project: "Team", Name: "fuda-tasks"}, "docs")
	s.host, s.accounts, s.client = server.URL, server.URL, server.Client()
	return s
}

func asUser() context.Context { return board.WithToken(context.Background(), "user-token") }

func TestListBoardsKeepsOnlyFudaRepositories(t *testing.T) {
	s := serve(t, &fakeHost{})
	got, err := s.listBoards(asUser())
	if err != nil || !slices.Equal(got, []string{"acme/Team/fuda-tasks"}) {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestReadWriteSendsOneEditCommitOnTheHead(t *testing.T) {
	f := &fakeHost{head: "head-1", blob: "blob-1", content: "status: todo\n"}
	s := serve(t, f)
	content, version, err := s.ReadFile(asUser(), "develop", "docs/board/tasks/a.md")
	if err != nil || string(content) != "status: todo\n" || version != "blob-1" {
		t.Fatalf("read: %q %q %v", content, version, err)
	}
	if err := s.WriteFile(asUser(), "develop", "docs/board/tasks/a.md", []byte("status: done\n"), version, "Move a"); err != nil {
		t.Fatal(err)
	}
	if len(f.pushes) != 1 {
		t.Fatalf("pushes: %d", len(f.pushes))
	}
	update := f.pushes[0]["refUpdates"].([]any)[0].(map[string]any)
	if update["name"] != "refs/heads/develop" || update["oldObjectId"] != "head-1" {
		t.Errorf("ref update: %v", update)
	}
}

func TestWriteFileReportsAFileChangedSinceTheRead(t *testing.T) {
	f := &fakeHost{head: "head-1", blob: "blob-2"}
	err := serve(t, f).WriteFile(asUser(), "develop", "docs/a.md", []byte("x"), "blob-1", "m")
	if !errors.Is(err, board.ErrChanged) || len(f.pushes) != 0 {
		t.Errorf("err %v, pushes %d", err, len(f.pushes))
	}
}

func TestWriteFileRetriesWhenOnlyTheBranchMoved(t *testing.T) {
	f := &fakeHost{head: "head-1", blob: "blob-1", moveHead: true}
	if err := serve(t, f).WriteFile(asUser(), "develop", "docs/a.md", []byte("x"), "blob-1", "m"); err != nil {
		t.Fatal(err)
	}
	update := f.pushes[0]["refUpdates"].([]any)[0].(map[string]any)
	if len(f.pushes) != 1 || update["oldObjectId"] != "head-2" {
		t.Errorf("pushes %v", f.pushes)
	}
}

func TestWriteFileFailsWhenThePushIsRefusedOnAnUnchangedBranch(t *testing.T) {
	f := &fakeHost{head: "head-1", blob: "blob-1", refuse: true}
	err := serve(t, f).WriteFile(asUser(), "develop", "docs/a.md", []byte("x"), "blob-1", "m")
	if err == nil || errors.Is(err, board.ErrChanged) {
		t.Errorf("err %v", err)
	}
}

func TestCanWriteChecksTheContributePermission(t *testing.T) {
	ok, err := serve(t, &fakeHost{}).CanWrite(asUser())
	if err != nil || !ok {
		t.Errorf("got %v, %v", ok, err)
	}
}

func TestCanWriteIsFalseWhenTheHostForbidsTheCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)
	s := New(Repo{Org: "acme", Project: "Team", Name: "fuda-tasks"}, "docs")
	s.host, s.client = server.URL, server.Client()
	ok, err := s.CanWrite(asUser())
	if err != nil || ok {
		t.Errorf("got %v, %v", ok, err)
	}
}

func TestLastEditorIsTheLatestCommitAuthor(t *testing.T) {
	who, err := serve(t, &fakeHost{}).LastEditor(asUser(), "develop", "docs/a.md")
	if err != nil || who != "Ben" {
		t.Errorf("got %q, %v", who, err)
	}
}

func TestCallsWithoutATokenNeedLogin(t *testing.T) {
	_, err := serve(t, &fakeHost{}).Head(context.Background(), "develop")
	if !errors.Is(err, board.ErrUnauthorized) {
		t.Errorf("err %v", err)
	}
}
