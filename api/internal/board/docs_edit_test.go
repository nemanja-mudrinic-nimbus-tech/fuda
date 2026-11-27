package board

import (
	"context"
	"errors"
	"testing"
)

const docPath = "docs/guide.md"

type messageHost struct {
	*fakeHost
	message string
}

func (m *messageHost) WriteFile(ctx context.Context, branch, path string, content []byte, version, message string) error {
	m.message = message
	return m.fakeHost.WriteFile(ctx, branch, path, content, version, message)
}

func newDocsHost() *messageHost {
	return &messageHost{fakeHost: newFakeHost(map[string]string{
		taskPath:                    taskContent("backlog", ""),
		"docs/board/archive/T-2.md": "---\nid: T-2\ntitle: Two\nstatus: done\n---\n",
		"docs/board/stages.md":      "---\nstages:\n  - name: Backlog\n    status: [backlog]\n---\n",
		docPath:                     "# Guide\n",
		"docs/adr/0001.md":          "# ADR\n",
	})}
}

func newDocsService(t *testing.T, host *messageHost) *Service {
	t.Helper()
	s := NewService(host, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "main"})
	if err := s.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSaveDocWritesTheContentAndRefreshesTheDoc(t *testing.T) {
	host := newDocsHost()
	s := newDocsService(t, host)
	ctx := context.Background()

	source, err := s.DocSource(ctx, "guide.md")
	if err != nil {
		t.Fatal(err)
	}
	check(t, "source", source.Content, "# Guide\n")

	if err := s.SaveDoc(ctx, SaveDocRequest{Path: "guide.md", Version: source.Version, Content: "# Guide\n\nMore.\n", Message: "  "}); err != nil {
		t.Fatal(err)
	}

	check(t, "file", string(host.files[docPath]), "# Guide\n\nMore.\n")
	check(t, "default message", host.message, "docs: edit guide.md")
	check(t, "writes", host.writes, 1)
	view, err := s.Doc("guide.md")
	if err != nil {
		t.Fatal(err)
	}
	if view.HTML == "" || view.Title != "Guide" {
		t.Errorf("doc not refreshed: %+v", view)
	}
	if got, _ := s.DocSource(ctx, "guide.md"); got.Content != "# Guide\n\nMore.\n" {
		t.Errorf("source after save: %q", got.Content)
	}
}

func TestSaveDocKeepsAGivenMessage(t *testing.T) {
	host := newDocsHost()
	s := newDocsService(t, host)

	if err := s.SaveDoc(context.Background(), SaveDocRequest{Path: "guide.md", Version: "0", Content: "x", Message: "docs: fix typo"}); err != nil {
		t.Fatal(err)
	}
	check(t, "message", host.message, "docs: fix typo")
}

func TestSaveDocConflictsWhenTheVersionMoved(t *testing.T) {
	host := newDocsHost()
	host.editor = "Ben"
	s := newDocsService(t, host)
	host.mu.Lock()
	host.files[docPath] = []byte("# Guide\n\nBen was here.\n")
	host.versions[docPath]++
	host.mu.Unlock()

	err := s.SaveDoc(context.Background(), SaveDocRequest{Path: "guide.md", Version: "0", Content: "mine"})

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("got %v, want a conflict", err)
	}
	check(t, "message", conflict.Message, "Ben changed this document since you started editing. Copy your changes, then reload the page.")
	check(t, "file", string(host.files[docPath]), "# Guide\n\nBen was here.\n")
	check(t, "writes", host.writes, 0)
}

func TestDocEditingRefusals(t *testing.T) {
	host := newDocsHost()
	s := newDocsService(t, host)
	ctx := context.Background()

	tests := []struct {
		name, path string
		want       error
	}{
		{"a task file", "board/tasks/T-1.md", ErrNotEditable},
		{"an archived task", "board/archive/T-2.md", ErrNotEditable},
		{"a board config file", "board/stages.md", ErrNotEditable},
		{"a path outside docs", "../README.md", ErrNotEditable},
		{"a path escaping through a doc folder", "adr/../../api/main.go", ErrNotEditable},
		{"a missing doc", "nope.md", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := s.DocSource(ctx, tt.path); !errors.Is(err, tt.want) {
				t.Errorf("source: got %v, want %v", err, tt.want)
			}
			err := s.SaveDoc(ctx, SaveDocRequest{Path: tt.path, Version: "0", Content: "x"})
			if !errors.Is(err, tt.want) {
				t.Errorf("save: got %v, want %v", err, tt.want)
			}
		})
	}
	check(t, "writes", host.writes, 0)
}

func TestReadOnlyAccountsCannotEditDocs(t *testing.T) {
	host := newDocsHost()
	host.readOnly = true
	s := newDocsService(t, host)
	ctx := context.Background()

	_, sourceErr := s.DocSource(ctx, "guide.md")
	saveErr := s.SaveDoc(ctx, SaveDocRequest{Path: "guide.md", Version: "0", Content: "x"})

	if !errors.Is(sourceErr, ErrForbidden) || !errors.Is(saveErr, ErrForbidden) {
		t.Fatalf("got %v and %v, want ErrForbidden", sourceErr, saveErr)
	}
	check(t, "writes", host.writes, 0)
}
