package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"fuda/internal/board"
	"fuda/internal/login"
)

type hostLogin struct {
	host      string
	header    string
	loggedOut *atomic.Int32
}

func (l hostLogin) Host() string { return l.host }

func (l hostLogin) Token(_ http.ResponseWriter, r *http.Request) (string, error) {
	if token := r.Header.Get(l.header); token != "" {
		return token, nil
	}
	return "", login.ErrNotLoggedIn
}

func (hostLogin) Routes(*http.ServeMux) {}

func (l hostLogin) Logout(http.ResponseWriter, *http.Request) { l.loggedOut.Add(1) }

type hostsServer struct {
	url       string
	loggedOut *atomic.Int32
}

func newHostsServer(t *testing.T) hostsServer {
	t.Helper()
	files := map[string][]byte{"docs/board/tasks/A-1.md": []byte("---\nid: A-1\ntitle: One\nstatus: backlog\n---\n")}
	boards := board.NewBoards(slog.New(slog.DiscardHandler), func(board.BoardID) (*board.Service, error) {
		return board.NewService(memberSource{files}, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, func(_ context.Context, host string) ([]board.BoardID, error) {
		switch host {
		case "github":
			return []board.BoardID{{Host: "github", Repo: "o/fuda-tasks"}}, nil
		case "azure":
			return []board.BoardID{{Host: "azure", Repo: "acme/Team/fuda-tasks"}}, nil
		}
		return nil, nil
	})
	loggedOut := &atomic.Int32{}
	server := httptest.NewServer(NewHandler(slog.New(slog.DiscardHandler), boards, fstest.MapFS{},
		hostLogin{"github", "X-Github-Token", loggedOut}, hostLogin{"azure", "X-Azure-Token", loggedOut}))
	t.Cleanup(server.Close)
	return hostsServer{server.URL, loggedOut}
}

type listing struct {
	Boards []struct{ Host, Path string }
	Hosts  []struct {
		Host     string
		LoggedIn bool
		Login    string
	}
}

func (h hostsServer) do(t *testing.T, method, path string, headers map[string]string) (int, listing) {
	t.Helper()
	req, _ := http.NewRequest(method, h.url+path, nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var body listing
	_ = json.NewDecoder(res.Body).Decode(&body)
	return res.StatusCode, body
}

func TestThePickerListsBoardsFromEveryHostTheUserIsLoggedInTo(t *testing.T) {
	server := newHostsServer(t)
	status, got := server.do(t, "GET", "/api/boards", map[string]string{"X-Github-Token": "member", "X-Azure-Token": "member"})
	if status != http.StatusOK || len(got.Boards) != 2 || got.Boards[0].Host != "github" || got.Boards[1].Host != "azure" {
		t.Fatalf("%d %+v", status, got)
	}
	for _, host := range got.Hosts {
		if !host.LoggedIn {
			t.Errorf("%s shown as logged out", host.Host)
		}
	}
}

func TestAHostWithoutALoginDoesNotHideTheOtherHost(t *testing.T) {
	server := newHostsServer(t)
	status, got := server.do(t, "GET", "/api/boards", map[string]string{"X-Github-Token": "member"})
	if status != http.StatusOK || len(got.Boards) != 1 || got.Boards[0].Host != "github" {
		t.Fatalf("%d %+v", status, got)
	}
	if len(got.Hosts) != 2 || !got.Hosts[0].LoggedIn {
		t.Fatalf("hosts: %+v", got.Hosts)
	}
	if azure := got.Hosts[1]; azure.Host != "azure" || azure.LoggedIn || azure.Login != "/auth/azure/login" {
		t.Errorf("azure: %+v", azure)
	}
}

func TestEachBoardChecksTheLoginOfItsOwnHost(t *testing.T) {
	server := newHostsServer(t)
	githubOnly := map[string]string{"X-Github-Token": "member"}
	if status, _ := server.do(t, "GET", "/api/github/o/fuda-tasks/board", githubOnly); status != http.StatusOK {
		t.Errorf("github board: %d", status)
	}
	if status, _ := server.do(t, "GET", "/api/azure/acme/Team/fuda-tasks/board", githubOnly); status != http.StatusUnauthorized {
		t.Errorf("azure board with a github login: %d", status)
	}
}

func TestLogOutOfAllClearsEveryHost(t *testing.T) {
	server := newHostsServer(t)
	if status, _ := server.do(t, "POST", "/auth/logout", nil); status != http.StatusNoContent {
		t.Fatalf("status %d", status)
	}
	if got := server.loggedOut.Load(); got != 2 {
		t.Errorf("hosts logged out: %d, want 2", got)
	}
}
