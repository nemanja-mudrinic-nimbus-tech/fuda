package board

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeHost struct {
	mu          sync.Mutex
	files       map[string][]byte
	versions    map[string]int
	head        int
	writes      int
	editor      string
	readOnly    bool
	beforeWrite func(h *fakeHost)
}

func newFakeHost(files map[string]string) *fakeHost {
	h := &fakeHost{files: map[string][]byte{}, versions: map[string]int{}}
	for p, content := range files {
		h.files[p] = []byte(content)
	}
	return h
}

func (h *fakeHost) Head(context.Context, string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return fmt.Sprint(h.head), nil
}

func (h *fakeHost) Files(context.Context, string) (map[string][]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := map[string][]byte{}
	for p, content := range h.files {
		out[p] = content
	}
	return out, nil
}

func (h *fakeHost) ReadFile(_ context.Context, _, path string) ([]byte, string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	content, ok := h.files[path]
	if !ok {
		return nil, "", ErrNotFound
	}
	return content, fmt.Sprint(h.versions[path]), nil
}

func (h *fakeHost) WriteFile(_ context.Context, _, path string, content []byte, version, _ string) error {
	h.mu.Lock()
	hook := h.beforeWrite
	h.mu.Unlock()
	if hook != nil {
		hook(h)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if version != fmt.Sprint(h.versions[path]) {
		return ErrChanged
	}
	h.files[path] = content
	h.versions[path]++
	h.head++
	h.writes++
	return nil
}

func (h *fakeHost) LastEditor(context.Context, string, string) (string, error) {
	return h.editor, nil
}

func (h *fakeHost) CanWrite(context.Context) (bool, error) {
	return !h.readOnly, nil
}

func (h *fakeHost) edit(change func(string) string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.files[taskPath] = []byte(change(string(h.files[taskPath])))
	h.versions[taskPath]++
	h.head++
}

func (h *fakeHost) file() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return string(h.files[taskPath])
}

const taskPath = "docs/board/tasks/T-1.md"

func taskContent(status string, extra string) string {
	return "---\nid: T-1\ntitle: One\nstatus: " + status + "\n" + extra + "---\n\nBody\n"
}

func newMoveService(t *testing.T, host *fakeHost) *Service {
	t.Helper()
	s := NewService(host, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
	s.now = func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	if err := s.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func cardColumn(t *testing.T, s *Service, id string) string {
	t.Helper()
	view, _ := s.Board()
	return card(view.Board, id).Column
}

func TestMoveWritesTheStatusAndClaimsTheTask(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "")})
	s := newMoveService(t, host)

	err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})
	if err != nil {
		t.Fatal(err)
	}

	check(t, "file", host.file(), taskContent("in progress", "claimed: 2026-10-09\n"))
	check(t, "writes", host.writes, 1)
	check(t, "column after the move", cardColumn(t, s, "T-1"), "in-progress")
}

func TestMoveClaimsOnlyWhenLeavingTheFirstStage(t *testing.T) {
	tests := []struct {
		name, status, extra, to, want string
	}{
		{"leaving a later stage adds no claim", "in progress", "", "testing", taskContent("testing", "")},
		{"a set claim stays", "in progress", "claimed: 2026-01-01\n", "backlog", taskContent("backlog", "claimed: 2026-01-01\n")},
		{"moving back and forth keeps the first claim", "backlog", "claimed: 2026-01-01\n", "in-progress", taskContent("in progress", "claimed: 2026-01-01\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := newFakeHost(map[string]string{taskPath: taskContent(tt.status, tt.extra)})
			s := newMoveService(t, host)
			if err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: tt.status, Column: tt.to}); err != nil {
				t.Fatal(err)
			}
			check(t, "file", host.file(), tt.want)
		})
	}
}

func TestMoveRetriesSilentlyWhenAnotherLineChanged(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "labels: a\n")})
	s := newMoveService(t, host)
	var once sync.Once
	host.beforeWrite = func(h *fakeHost) {
		once.Do(func() {
			h.edit(func(c string) string { return strings.Replace(c, "labels: a", "labels: a, b", 1) })
		})
	}

	if err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"}); err != nil {
		t.Fatal(err)
	}

	check(t, "file", host.file(), taskContent("in progress", "claimed: 2026-10-09\nlabels: a, b\n"))
}

func TestMoveLosesWhenStatusChangedFirst(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "")})
	host.editor = "Ben"
	s := newMoveService(t, host)
	var once sync.Once
	host.beforeWrite = func(h *fakeHost) {
		once.Do(func() {
			h.edit(func(c string) string { return strings.Replace(c, "status: backlog", "status: testing", 1) })
		})
	}

	err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("got %v, want a conflict", err)
	}
	check(t, "message", conflict.Message, "Ben moved this to Testing just now")
	check(t, "file", host.file(), taskContent("testing", ""))
	check(t, "writes", host.writes, 0)
}

func TestMoveLosesWhenStatusChangedBeforeTheRead(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "")})
	s := newMoveService(t, host)
	host.edit(func(c string) string { return strings.Replace(c, "status: backlog", "status: testing", 1) })

	err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("got %v, want a conflict", err)
	}
	check(t, "message", conflict.Message, "Someone moved this to Testing just now")
	check(t, "file", host.file(), taskContent("testing", ""))
}

func TestMoveGivesUpInsteadOfForcing(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "labels: 0\n")})
	s := newMoveService(t, host)
	n := 0
	host.beforeWrite = func(h *fakeHost) {
		n++
		h.edit(func(c string) string {
			return strings.Replace(c, fmt.Sprintf("labels: %d", n-1), fmt.Sprintf("labels: %d", n), 1)
		})
	}

	err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("got %v, want a conflict", err)
	}
	check(t, "writes", host.writes, 0)
	if !strings.Contains(host.file(), "status: backlog") {
		t.Errorf("fuda overwrote a file it lost: %q", host.file())
	}
}

func TestMoveRefusals(t *testing.T) {
	archivePath := "docs/board/archive/T-2.md"
	host := newFakeHost(map[string]string{
		taskPath:    taskContent("backlog", ""),
		archivePath: "---\nid: T-2\ntitle: Two\nstatus: done\n---\n",
	})
	s := newMoveService(t, host)
	ctx := context.Background()

	tests := []struct {
		name string
		req  MoveRequest
		want error
	}{
		{"an archived task", MoveRequest{TaskID: "T-2", Seen: "done", Column: "backlog"}, ErrLocked},
		{"a stage that only open PRs fill", MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-review"}, ErrLocked},
		{"an unknown task", MoveRequest{TaskID: "T-9", Seen: "backlog", Column: "backlog"}, ErrNotFound},
		{"an unknown stage", MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "nowhere"}, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Move(ctx, tt.req); !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
	check(t, "writes", host.writes, 0)
}

func TestMoveNeedsASourceThatWrites(t *testing.T) {
	s := NewService(fakeSource{files: map[string][]byte{taskPath: []byte(taskContent("backlog", ""))}}, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
	if err := s.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("got %v, want ErrForbidden", err)
	}
}

func TestReadOnlyAccountsCannotMoveOrAssign(t *testing.T) {
	host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "")})
	host.readOnly = true
	s := newMoveService(t, host)

	moveErr := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "backlog", Column: "in-progress"})
	assignErr := s.Assign(context.Background(), AssignRequest{TaskID: "T-1", Owners: []string{"Ann"}})

	if !errors.Is(moveErr, ErrForbidden) || !errors.Is(assignErr, ErrForbidden) {
		t.Fatalf("got %v and %v, want ErrForbidden", moveErr, assignErr)
	}
	check(t, "file", host.file(), taskContent("backlog", ""))
	check(t, "writes", host.writes, 0)
}

func TestServiceReportsWriteAccess(t *testing.T) {
	for _, readOnly := range []bool{false, true} {
		host := newFakeHost(map[string]string{taskPath: taskContent("backlog", "")})
		host.readOnly = readOnly
		s := newMoveService(t, host)

		can, err := s.CanWrite(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		check(t, fmt.Sprintf("can write when readOnly=%v", readOnly), can, !readOnly)
	}
}

type countingHost struct {
	*fakeHost
	calls int
}

func (c *countingHost) CanWrite(ctx context.Context) (bool, error) {
	c.calls++
	return c.fakeHost.CanWrite(ctx)
}

func TestWriteAccessIsKeptForAMinutePerToken(t *testing.T) {
	host := &countingHost{fakeHost: newFakeHost(map[string]string{taskPath: taskContent("backlog", "")})}
	s := NewService(host, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	ctx := WithToken(context.Background(), "ann")

	for range 3 {
		if _, err := s.CanWrite(ctx); err != nil {
			t.Fatal(err)
		}
	}
	check(t, "checks within a minute", host.calls, 1)

	if _, err := s.CanWrite(WithToken(context.Background(), "ben")); err != nil {
		t.Fatal(err)
	}
	check(t, "a second token is checked on its own", host.calls, 2)

	now = now.Add(2 * time.Minute)
	if _, err := s.CanWrite(ctx); err != nil {
		t.Fatal(err)
	}
	check(t, "checks after a minute", host.calls, 3)
}
