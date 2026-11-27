package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"fuda/internal/board"
)

const codeRepoTask = "---\nid: FU-1\ntitle: First\nstatus: backlog\n---\n\n# FU-1: First\n"

type codeRepoHost struct {
	mu      sync.Mutex
	task    string
	written string
	paths   []string
}

func (h *codeRepoHost) zipball(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range map[string]string{
		"o-app-abc/docs/board/tasks/epic/FU-1-first.md": h.task,
		"o-app-abc/docs/guide.md":                       "# Guide\n",
		"o-app-abc/README.md":                           "# App\n",
		"o-app-abc/main.go":                             "package main\n",
	} {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func (h *codeRepoHost) serve(t *testing.T) http.HandlerFunc {
	const taskPath = "/repos/o/app/contents/docs/board/tasks/epic/FU-1-first.md"
	return func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		defer h.mu.Unlock()
		switch {
		case r.URL.Path == "/repos/o/app/branches/main":
			_, _ = w.Write([]byte(`{"commit":{"sha":"abc"}}`))
		case r.URL.Path == "/repos/o/app/zipball/main":
			_, _ = w.Write(h.zipball(t))
		case r.URL.Path == "/repos/o/app":
			_, _ = w.Write([]byte(`{"permissions":{"push":true}}`))
		case r.URL.Path == taskPath && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "sha": "s1", "content": base64.StdEncoding.EncodeToString([]byte(h.task))})
		case r.URL.Path == taskPath && r.Method == http.MethodPut:
			var payload struct{ Content, Branch string }
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &payload)
			content, _ := base64.StdEncoding.DecodeString(payload.Content)
			h.task = string(content)
			h.written = payload.Branch
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
		h.paths = append(h.paths, r.Method+" "+r.URL.Path)
	}
}

func TestListedCodeRepositoryWithDocsBoardSyncsAndMoves(t *testing.T) {
	host := &codeRepoHost{task: codeRepoTask}
	server := httptest.NewServer(host.serve(t))
	defer server.Close()
	source := New("o/app", "docs")
	source.base = server.URL
	source.client = server.Client()
	service := board.NewService(source, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "main", InCodeRepo: true})
	ctx := board.WithToken(context.Background(), "t")

	if err := service.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	view, ok := service.Board()
	if !ok || len(view.Cards) != 1 || view.Cards[0].ID != "FU-1" || len(view.Problems) != 0 {
		t.Fatalf("board: %+v, %v", view, ok)
	}
	if docs := service.Docs(); !slices.Equal(docs, []string{"guide.md"}) {
		t.Errorf("docs: %v", docs)
	}

	err := service.Move(ctx, board.MoveRequest{TaskID: "FU-1", FromStatus: "backlog", ToColumn: "in-progress"})
	if err != nil {
		t.Fatal(err)
	}
	if host.written != "main" || !strings.Contains(host.task, `status: in progress`) {
		t.Errorf("branch %q, file %q", host.written, host.task)
	}
}
