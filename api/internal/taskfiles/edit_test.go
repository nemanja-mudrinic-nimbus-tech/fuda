package taskfiles

import (
	"errors"
	"testing"
)

func TestApplyMove(t *testing.T) {
	tests := []struct {
		name string
		in   string
		move Move
		want string
	}{
		{
			name: "changes only the status line",
			in:   "---\nid: T-1\ntitle: One\nstatus: backlog\nlabels: a, b\n---\n\nBody: keep\n",
			move: Move{Status: "in progress"},
			want: "---\nid: T-1\ntitle: One\nstatus: in progress\nlabels: a, b\n---\n\nBody: keep\n",
		},
		{
			name: "adds claimed right after status",
			in:   "---\nid: T-1\ntitle: One\nstatus: backlog\nadded: 2026-01-01\n---\nBody\n",
			move: Move{Status: "in progress", Claimed: "2026-10-09"},
			want: "---\nid: T-1\ntitle: One\nstatus: in progress\nclaimed: 2026-10-09\nadded: 2026-01-01\n---\nBody\n",
		},
		{
			name: "never changes a claimed date that is set",
			in:   "---\nid: T-1\ntitle: One\nstatus: testing\nclaimed: 2026-02-02\n---\nBody\n",
			move: Move{Status: "in progress", Claimed: "2026-10-09"},
			want: "---\nid: T-1\ntitle: One\nstatus: in progress\nclaimed: 2026-02-02\n---\nBody\n",
		},
		{
			name: "fills an empty claimed line",
			in:   "---\nid: T-1\ntitle: One\nstatus: backlog\nclaimed:\n---\n",
			move: Move{Status: "in progress", Claimed: "2026-10-09"},
			want: "---\nid: T-1\ntitle: One\nstatus: in progress\nclaimed: 2026-10-09\n---\n",
		},
		{
			name: "keeps windows line endings and the byte order mark",
			in:   "\xef\xbb\xbf---\r\nid: T-1\r\ntitle: One\r\nstatus: backlog\r\n---\r\nBody\r\n",
			move: Move{Status: "in progress", Claimed: "2026-10-09"},
			want: "\xef\xbb\xbf---\r\nid: T-1\r\ntitle: One\r\nstatus: in progress\r\nclaimed: 2026-10-09\r\n---\r\nBody\r\n",
		},
		{
			name: "quotes a status that needs it",
			in:   "---\nid: T-1\ntitle: One\nstatus: backlog\n---\n",
			move: Move{Status: "in: review"},
			want: "---\nid: T-1\ntitle: One\nstatus: 'in: review'\n---\n",
		},
		{
			name: "ignores status lines in the body and nested keys",
			in:   "---\nid: T-1\ntitle: One\nmeta:\n  status: nested\nstatus: backlog\n---\nstatus: body\n",
			move: Move{Status: "testing"},
			want: "---\nid: T-1\ntitle: One\nmeta:\n  status: nested\nstatus: testing\n---\nstatus: body\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyMove([]byte(tt.in), tt.move)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestApplyMoveNeedsAStatusLine(t *testing.T) {
	for name, in := range map[string]string{
		"no frontmatter":      "Body only\n",
		"no status line":      "---\nid: T-1\ntitle: One\n---\n",
		"status only in body": "---\nid: T-1\n---\nstatus: backlog\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ApplyMove([]byte(in), Move{Status: "testing"}); !errors.Is(err, ErrNoStatusLine) {
				t.Errorf("got %v, want ErrNoStatusLine", err)
			}
		})
	}
}

func TestParseOne(t *testing.T) {
	task, err := ParseOne("docs/board/tasks/T-1.md", []byte("---\nid: T-1\ntitle: One\nstatus: backlog\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != "backlog" || task.ID != "T-1" {
		t.Errorf("got %+v", task)
	}
	if _, err := ParseOne("docs/board/tasks/T-1.md", []byte("nope")); err == nil {
		t.Error("want an error for a file without frontmatter")
	}
}

func TestApplyAssign(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		owners []string
		want   string
	}{
		{
			name:   "rewrites comma text",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Ann, Ben\nlabels: a\n---\nBody\n",
			owners: []string{"Cy"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Cy\nlabels: a\n---\nBody\n",
		},
		{
			name:   "joins several names in comma text",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Ann\n---\n",
			owners: []string{"Ann", "Ben"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Ann, Ben\n---\n",
		},
		{
			name:   "keeps a block list",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner:\n  - Ann\n  - Ben\nlabels: a\n---\n",
			owners: []string{"Ben", "Cy", "Di"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner:\n  - Ben\n  - Cy\n  - Di\nlabels: a\n---\n",
		},
		{
			name:   "keeps a block list without indent",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner:\n- Ann\nlabels: a\n---\n",
			owners: []string{"Cy"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner:\n- Cy\nlabels: a\n---\n",
		},
		{
			name:   "keeps a flow list",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: [Ann, Ben]\n---\n",
			owners: []string{"Cy"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: [Cy]\n---\n",
		},
		{
			name:   "adds a new owner line as comma text after status",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nclaimed: 2026-01-01\n---\nBody\n",
			owners: []string{"Ann", "Ben"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Ann, Ben\nclaimed: 2026-01-01\n---\nBody\n",
		},
		{
			name:   "fills an empty owner line",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner:\nlabels: a\n---\n",
			owners: []string{"Ann"},
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Ann\nlabels: a\n---\n",
		},
		{
			name:   "deletes a comma text line when no owner is left",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner: Ann\nlabels: a\n---\n",
			owners: nil,
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nlabels: a\n---\n",
		},
		{
			name:   "deletes a block list with its items",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\nowner:\n  - Ann\n  - Ben\nlabels: a\n---\n",
			owners: nil,
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\nlabels: a\n---\n",
		},
		{
			name:   "adds nothing when there is no owner and none is asked",
			in:     "---\nid: T-1\ntitle: One\nstatus: backlog\n---\n",
			owners: nil,
			want:   "---\nid: T-1\ntitle: One\nstatus: backlog\n---\n",
		},
		{
			name:   "keeps windows line endings",
			in:     "---\r\nid: T-1\r\ntitle: One\r\nstatus: backlog\r\nowner: Ann\r\n---\r\nBody\r\n",
			owners: []string{"Ben"},
			want:   "---\r\nid: T-1\r\ntitle: One\r\nstatus: backlog\r\nowner: Ben\r\n---\r\nBody\r\n",
		},
		{
			name:   "ignores owner lines in the body and nested keys",
			in:     "---\nid: T-1\ntitle: One\nmeta:\n  owner: nested\nstatus: backlog\n---\nowner: body\n",
			owners: []string{"Ann"},
			want:   "---\nid: T-1\ntitle: One\nmeta:\n  owner: nested\nstatus: backlog\nowner: Ann\n---\nowner: body\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyAssign([]byte(tt.in), tt.owners)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestApplyAssignNeedsAStatusLine(t *testing.T) {
	_, err := ApplyAssign([]byte("---\nid: T-1\n---\n"), []string{"Ann"})
	if !errors.Is(err, ErrNoStatusLine) {
		t.Errorf("got %v, want ErrNoStatusLine", err)
	}
}
