package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"fuda/internal/board"
	"fuda/internal/login"
	"fuda/internal/source/local"
)

func writeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for p, content := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestAPISmoke(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"docs/board/tasks/SS-1-login.md": "---\nid: SS-1\ntitle: Login\nstatus: in progress\nowner: Ana\nlabels: [\"type:bug\"]\n---\nNeeds [the rules](../TASKS.md) and SS-2.\n",
		"docs/board/tasks/SS-2.md":       "---\nid: SS-2\ntitle: Logout\nstatus: backlog\n---\nRotate the session key.\n",
		"docs/board/tasks/broken.md":     "no frontmatter\n",
		"docs/board/archive/SS-0.md":     "---\nid: SS-0\ntitle: Old\nstatus: done\ndone: 2026-09-01\n---\n",
		"docs/board/TASKS.md":            "# Task rules\n\nClaim first.\n",
		"docs/board/assets/flow.png":     "\x89PNG fake",
	})
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), testBoards(t, map[string]string{"test": root}), fstest.MapFS{}))
	defer server.Close()
	base := server.URL + "/api/local/test"

	var b struct {
		Title    string
		Columns  []struct{ ID string }
		Cards    []struct{ ID, Column string }
		Problems []struct{ Path string }
	}
	getJSON(t, base+"/board", http.StatusOK, &b)
	if b.Title != "test" || len(b.Columns) != 6 || len(b.Cards) != 2 || len(b.Problems) != 1 {
		t.Fatalf("board: %+v", b)
	}

	var task struct{ ID, HTML string }
	getJSON(t, base+"/tasks/SS-1", http.StatusOK, &task)
	if !strings.Contains(task.HTML, `href="/local/test/docs/board/TASKS.md"`) {
		t.Errorf("task html: %s", task.HTML)
	}
	getJSON(t, base+"/tasks/NOPE", http.StatusNotFound, nil)

	var ids []string
	getJSON(t, base+"/search?q=session", http.StatusOK, &ids)
	if len(ids) != 1 || ids[0] != "SS-2" {
		t.Errorf("search: %v", ids)
	}

	var archive []struct{ ID, Done string }
	getJSON(t, base+"/archive", http.StatusOK, &archive)
	if len(archive) != 1 || archive[0].Done != "2026-09-01" {
		t.Errorf("archive: %+v", archive)
	}

	var doc struct {
		Title     string
		Backlinks []string
	}
	getJSON(t, base+"/docs?path=board/TASKS.md", http.StatusOK, &doc)
	if doc.Title != "Task rules" || len(doc.Backlinks) != 1 {
		t.Errorf("doc: %+v", doc)
	}
	getJSON(t, base+"/docs?path=../../etc/passwd", http.StatusNotFound, nil)

	res, err := http.Get(base + "/files?path=docs/board/assets/flow.png")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || res.Header.Get("Content-Type") != "image/png" {
		t.Errorf("asset: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	getJSON(t, base+"/files?path=docs/board/TASKS.md", http.StatusNotFound, nil)

	for _, want := range []int{http.StatusAccepted, http.StatusTooManyRequests} {
		res, err := http.Post(base+"/sync", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != want {
			t.Errorf("sync: got %d, want %d", res.StatusCode, want)
		}
	}
}

func testBoards(t *testing.T, folders map[string]string) *board.Boards {
	t.Helper()
	return board.NewBoards(slog.New(slog.DiscardHandler), func(id board.BoardID) (*board.Service, error) {
		root, ok := folders[id.Repo]
		if !ok || id.Host != "local" {
			return nil, board.ErrNotFound
		}
		return board.NewService(local.New(root, "docs", "docs/board", "develop"), board.Options{
			Title: id.Repo, DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main",
			Cooldown: time.Minute, Origin: board.Origin{Host: "local", Repo: id.Repo, Path: id.Path()},
		}), nil
	}, func(context.Context, string) ([]board.BoardID, error) {
		var ids []board.BoardID
		for folder := range folders {
			ids = append(ids, board.BoardID{Host: "local", Repo: folder})
		}
		return ids, nil
	})
}

func TestBoardsAreServedByPath(t *testing.T) {
	one := writeRepo(t, map[string]string{"docs/board/tasks/ONE-1.md": "---\nid: ONE-1\ntitle: First\nstatus: backlog\n---\nSee [rules](../TASKS.md).\n", "docs/board/TASKS.md": "# Rules\n"})
	two := writeRepo(t, map[string]string{"docs/board/tasks/TWO-1.md": "---\nid: TWO-1\ntitle: Second\nstatus: backlog\n---\n"})
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), testBoards(t, map[string]string{"one": one, "two": two}), fstest.MapFS{}))
	defer server.Close()

	for folder, card := range map[string]string{"one": "ONE-1", "two": "TWO-1"} {
		var b struct{ Cards []struct{ ID string } }
		getJSON(t, server.URL+"/api/local/"+folder+"/board", http.StatusOK, &b)
		if len(b.Cards) != 1 || b.Cards[0].ID != card {
			t.Errorf("%s: %+v", folder, b.Cards)
		}
	}
	getJSON(t, server.URL+"/api/local/two/tasks/ONE-1", http.StatusNotFound, nil)

	var task struct{ HTML string }
	getJSON(t, server.URL+"/api/local/one/tasks/ONE-1", http.StatusOK, &task)
	if !strings.Contains(task.HTML, `href="/local/one/docs/board/TASKS.md"`) {
		t.Errorf("links are not prefixed with the Board path: %s", task.HTML)
	}

	getJSON(t, server.URL+"/api/local/nothing/board", http.StatusNotFound, nil)
	getJSON(t, server.URL+"/api/github/nobody/none/board", http.StatusNotFound, nil)
}

func getJSON(t *testing.T, url string, status int, v any) {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != status {
		t.Fatalf("%s: status %d, want %d: %s", url, res.StatusCode, status, body)
	}
	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			t.Fatalf("%s: %v", url, err)
		}
	}
}

type headerLogin struct{}

func (headerLogin) Token(_ http.ResponseWriter, r *http.Request) (string, error) {
	if token := r.Header.Get("X-Test-Token"); token != "" {
		return token, nil
	}
	return "", login.ErrNotLoggedIn
}

func (headerLogin) Routes(*http.ServeMux) {}

func (headerLogin) Logout(http.ResponseWriter, *http.Request) {}

func (headerLogin) Host() string { return "github" }

type azureHeaderLogin struct{ headerLogin }

func (azureHeaderLogin) Host() string { return "azure" }

func TestAnAzureLoginServesAzureBoardsOnly(t *testing.T) {
	files := map[string][]byte{"docs/board/tasks/A-1.md": []byte("---\nid: A-1\ntitle: One\nstatus: backlog\n---\n")}
	boards := board.NewBoards(slog.New(slog.DiscardHandler), func(id board.BoardID) (*board.Service, error) {
		return board.NewService(memberSource{files}, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, func(context.Context, string) ([]board.BoardID, error) {
		return []board.BoardID{{Host: "azure", Repo: "acme/Team/fuda-tasks"}}, nil
	})
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), boards, fstest.MapFS{}, azureHeaderLogin{}))
	defer server.Close()

	status := func(path string) int {
		req, _ := http.NewRequest(http.MethodGet, server.URL+path, nil)
		req.Header.Set("X-Test-Token", "member")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}
	if got := status("/api/azure/acme/Team/fuda-tasks/board"); got != http.StatusOK {
		t.Errorf("azure board: %d", got)
	}
	if got := status("/api/github/o/fuda-tasks/board"); got != http.StatusNotFound {
		t.Errorf("github board on an azure login: %d", got)
	}
}

type memberSource struct{ files map[string][]byte }

func (m memberSource) Head(ctx context.Context, _ string) (string, error) {
	switch board.TokenFrom(ctx) {
	case "member":
		return "head", nil
	case "banned":
		return "", board.ErrForbidden
	}
	return "", board.ErrBranchMissing
}

func (m memberSource) Files(context.Context, string) (map[string][]byte, error) { return m.files, nil }

func TestGitHubBoardsNeedALogin(t *testing.T) {
	files := map[string][]byte{"docs/board/tasks/A-1.md": []byte("---\nid: A-1\ntitle: One\nstatus: backlog\n---\n")}
	boards := board.NewBoards(slog.New(slog.DiscardHandler), func(id board.BoardID) (*board.Service, error) {
		return board.NewService(memberSource{files}, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, func(ctx context.Context, _ string) ([]board.BoardID, error) {
		if board.TokenFrom(ctx) != "member" {
			return nil, board.ErrForbidden
		}
		return []board.BoardID{{Host: "github", Repo: "o/fuda-tasks"}}, nil
	})
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), boards, fstest.MapFS{}, headerLogin{}))
	defer server.Close()

	get := func(path, token string, status int, v any) {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, server.URL+path, nil)
		if token != "" {
			req.Header.Set("X-Test-Token", token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = res.Body.Close() }()
		if res.StatusCode != status {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("%s as %q: status %d, want %d: %s", path, token, res.StatusCode, status, body)
		}
		if v != nil {
			if err := json.NewDecoder(res.Body).Decode(v); err != nil {
				t.Fatal(err)
			}
		}
	}

	get("/api/github/o/fuda-tasks/board", "", http.StatusUnauthorized, nil)

	var b struct{ Cards []struct{ ID string } }
	get("/api/github/o/fuda-tasks/board", "member", http.StatusOK, &b)
	if len(b.Cards) != 1 {
		t.Errorf("cards: %+v", b.Cards)
	}
	get("/api/github/o/fuda-tasks/board", "stranger", http.StatusNotFound, nil)
	get("/api/github/o/fuda-tasks/board", "banned", http.StatusForbidden, nil)

	var listing struct {
		Boards []struct{ Host, Repo, Path, Title string }
	}
	get("/api/boards", "member", http.StatusOK, &listing)
	if len(listing.Boards) != 1 || listing.Boards[0].Path != "/github/o/fuda-tasks" || listing.Boards[0].Title != "fuda-tasks" {
		t.Errorf("listing: %+v", listing)
	}
}

type movableSource struct {
	mu      sync.Mutex
	content string
	version int
	editor  string
}

func (m *movableSource) Head(context.Context, string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m.version), nil
}

func (m *movableSource) Files(context.Context, string) (map[string][]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string][]byte{"docs/board/tasks/A-1.md": []byte(m.content)}, nil
}

func (m *movableSource) ReadFile(ctx context.Context, _, _ string) ([]byte, string, error) {
	if board.TokenFrom(ctx) == "" {
		return nil, "", board.ErrUnauthorized
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return []byte(m.content), strconv.Itoa(m.version), nil
}

func (m *movableSource) WriteFile(ctx context.Context, _, _ string, content []byte, version, _ string) error {
	if board.TokenFrom(ctx) == "" {
		return board.ErrUnauthorized
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if version != strconv.Itoa(m.version) {
		return board.ErrChanged
	}
	m.content = string(content)
	m.version++
	return nil
}

func (m *movableSource) LastEditor(context.Context, string, string) (string, error) {
	return m.editor, nil
}

func TestMoveThroughTheHandler(t *testing.T) {
	source := &movableSource{content: "---\nid: A-1\ntitle: One\nstatus: backlog\n---\n", editor: "Ben"}
	boards := board.NewBoards(slog.New(slog.DiscardHandler), func(board.BoardID) (*board.Service, error) {
		return board.NewService(source, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, nil)
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), boards, fstest.MapFS{}, headerLogin{}))
	defer server.Close()

	move := func(token, contentType, body string, status int) string {
		t.Helper()
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/github/o/fuda-tasks/tasks/A-1/move", strings.NewReader(body))
		req.Header.Set("Content-Type", contentType)
		if token != "" {
			req.Header.Set("X-Test-Token", token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = res.Body.Close() }()
		out, _ := io.ReadAll(res.Body)
		if res.StatusCode != status {
			t.Fatalf("status %d, want %d: %s", res.StatusCode, status, out)
		}
		return string(out)
	}
	const json = "application/json"

	move("", json, `{"seen":"backlog","column":"in-progress"}`, http.StatusUnauthorized)
	move("member", "text/plain", `{"seen":"backlog","column":"in-progress"}`, http.StatusUnsupportedMediaType)
	move("member", json, `nonsense`, http.StatusBadRequest)
	move("member", json, `{"seen":"backlog","column":"in-review"}`, http.StatusForbidden)

	move("member", json, `{"seen":"backlog","column":"in-progress"}`, http.StatusNoContent)
	if !strings.Contains(source.content, "status: in progress\n") {
		t.Errorf("file after the move: %q", source.content)
	}

	body := move("member", json, `{"seen":"backlog","column":"testing"}`, http.StatusConflict)
	if !strings.Contains(body, "Ben moved this to In progress just now") {
		t.Errorf("conflict body: %s", body)
	}
}

func TestAssignThroughTheHandler(t *testing.T) {
	source := &movableSource{content: "---\nid: A-1\ntitle: One\nstatus: backlog\n---\n", editor: "Ben"}
	boards := board.NewBoards(slog.New(slog.DiscardHandler), func(board.BoardID) (*board.Service, error) {
		return board.NewService(source, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, nil)
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), boards, fstest.MapFS{}, headerLogin{}))
	defer server.Close()

	assign := func(token, body string, status int) string {
		t.Helper()
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/github/o/fuda-tasks/tasks/A-1/assign", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("X-Test-Token", token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = res.Body.Close() }()
		out, _ := io.ReadAll(res.Body)
		if res.StatusCode != status {
			t.Fatalf("status %d, want %d: %s", res.StatusCode, status, out)
		}
		return string(out)
	}

	assign("", `{"seen":[],"owners":["Ann"]}`, http.StatusUnauthorized)
	assign("member", `{"seen":[],"owners":["Ann"]}`, http.StatusNoContent)
	if !strings.Contains(source.content, "owner: Ann\n") {
		t.Errorf("file after the assign: %q", source.content)
	}

	body := assign("member", `{"seen":[],"owners":["Cy"]}`, http.StatusConflict)
	if !strings.Contains(body, "Ben set the owners to Ann just now") {
		t.Errorf("conflict body: %s", body)
	}
}
