package httpapi

import (
	"context"
	"encoding/json"
	"errors"
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
	host      board.Host
	header    string
	loggedOut *atomic.Int32
}

func (l hostLogin) Host() board.Host { return l.host }

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

func newHostsServer(t *testing.T, installs map[board.Host]board.AppInstall) hostsServer {
	t.Helper()
	files := map[string][]byte{"docs/board/tasks/A-1.md": []byte("---\nid: A-1\ntitle: One\nstatus: backlog\n---\n")}
	boards := board.NewRegistry(slog.New(slog.DiscardHandler), func(board.ID) (*board.Service, error) {
		return board.NewService(memberSource{files}, board.Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"}), nil
	}, func(_ context.Context, host board.Host) ([]board.ID, error) {
		switch host {
		case board.HostGitHub:
			return []board.ID{{Host: board.HostGitHub, Repo: "o/fuda-tasks"}}, nil
		case board.HostAzure:
			return []board.ID{{Host: board.HostAzure, Repo: "acme/Team/fuda-tasks"}}, nil
		}
		return nil, nil
	}, installs)
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

		InstallURL   string
		AppInstalled *bool
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
	server := newHostsServer(t, nil)
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
	server := newHostsServer(t, nil)
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
	server := newHostsServer(t, nil)
	githubOnly := map[string]string{"X-Github-Token": "member"}
	if status, _ := server.do(t, "GET", "/api/github/o/fuda-tasks/board", githubOnly); status != http.StatusOK {
		t.Errorf("github board: %d", status)
	}
	if status, _ := server.do(t, "GET", "/api/azure/acme/Team/fuda-tasks/board", githubOnly); status != http.StatusUnauthorized {
		t.Errorf("azure board with a github login: %d", status)
	}
}

func TestLogOutOfAllClearsEveryHost(t *testing.T) {
	server := newHostsServer(t, nil)
	if status, _ := server.do(t, "POST", "/auth/logout", nil); status != http.StatusNoContent {
		t.Fatalf("status %d", status)
	}
	if got := server.loggedOut.Load(); got != 2 {
		t.Errorf("hosts logged out: %d, want 2", got)
	}
}

func TestHostStatusSaysWhetherTheAppIsInstalled(t *testing.T) {
	for _, tc := range []struct {
		name      string
		installed func(context.Context) (bool, error)
		want      *bool
	}{
		{"installed", func(context.Context) (bool, error) { return true, nil }, new(true)},
		{"not installed", func(context.Context) (bool, error) { return false, nil }, new(false)},
		{"call failed", func(context.Context) (bool, error) { return false, errors.New("boom") }, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newHostsServer(t, map[board.Host]board.AppInstall{board.HostGitHub: {URL: "https://github.com/apps/fuda/installations/new", Installed: tc.installed}})
			_, got := server.do(t, "GET", "/api/boards", map[string]string{"X-Github-Token": "member", "X-Azure-Token": "member"})
			github, azure := got.Hosts[0], got.Hosts[1]
			if github.InstallURL != "https://github.com/apps/fuda/installations/new" {
				t.Errorf("install url: %q", github.InstallURL)
			}
			if (github.AppInstalled == nil) != (tc.want == nil) || (tc.want != nil && *github.AppInstalled != *tc.want) {
				t.Errorf("appInstalled: %v, want %v", github.AppInstalled, tc.want)
			}
			if azure.AppInstalled != nil || azure.InstallURL != "" {
				t.Errorf("azure: %+v", azure)
			}
		})
	}
}
