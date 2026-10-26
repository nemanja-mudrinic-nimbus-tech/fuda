package taskfiles

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var layout = Layout{BoardDir: "docs/board"}

func taskFile(front, body string) []byte {
	return []byte("---\n" + front + "\n---\n" + body)
}

func TestParseTaskShapes(t *testing.T) {
	files := map[string][]byte{
		"docs/board/tasks/S1-guard.md": taskFile(`id: "S1"
title: "guard the routers"
status: "in progress"
blocked_by: "Q26, J1"
owner: "Ron, Chinmay"
tester: ""
batch: 3
labels: ["theme:security", "area:be", "loose"]
size: "S"
added: "2026-09-25"
pr: [1944, "#1950", "chore/c1-ruff"]
done: ""`, "# S1\n\n## Done when\nit is.\n"),
		"docs/board/tasks/SS-2.md": taskFile(`id: SS-2
title: plain YAML
status: backlog
owner: [Nemanja, "Filip, Milan"]
labels:
  - type:bug
pr: chore/branch-name`, ""),
	}

	r := Parse(files, layout)
	if len(r.Problems) != 0 {
		t.Fatalf("unexpected problems: %+v", r.Problems)
	}
	if len(r.Tasks) != 2 {
		t.Fatalf("want 2 tasks, got %d", len(r.Tasks))
	}

	s1 := r.Tasks[0]
	check(t, "id", s1.ID, "S1")
	check(t, "status", s1.Status, "in progress")
	check(t, "blocked_by", s1.BlockedBy, "Q26, J1")
	checkDeep(t, "owners", s1.Owners, []string{"Ron", "Chinmay"})
	checkDeep(t, "testers", s1.Testers, []string(nil))
	checkDeep(t, "labels", s1.Labels, []Label{{"theme", "security"}, {"area", "be"}, {"", "loose"}})
	checkDeep(t, "prs", s1.PRs, []int{1944, 1950})
	check(t, "pr ref", s1.PRRef, "chore/c1-ruff")
	checkDeep(t, "custom", s1.Custom, []Field{{"batch", "3"}, {"size", "S"}})
	check(t, "body", s1.Body, "# S1\n\n## Done when\nit is.\n")

	ss2 := r.Tasks[1]
	checkDeep(t, "owners", ss2.Owners, []string{"Nemanja", "Filip", "Milan"})
	checkDeep(t, "labels", ss2.Labels, []Label{{"type", "bug"}})
	check(t, "pr ref", ss2.PRRef, "chore/branch-name")
}

func TestParseProblems(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		file   []byte
		reason string
	}{
		{"no frontmatter", "docs/board/tasks/A1.md", []byte("# A1\n"), "no frontmatter"},
		{"unclosed", "docs/board/tasks/A1.md", []byte("---\nid: A1\n"), "not closed"},
		{"broken yaml", "docs/board/tasks/A1.md", taskFile("id: [A1", ""), "invalid YAML"},
		{"missing id", "docs/board/tasks/A1.md", taskFile("title: t\nstatus: backlog", ""), "missing required field id"},
		{"missing title", "docs/board/tasks/A1.md", taskFile("id: A1\nstatus: backlog", ""), "missing required field title"},
		{"missing status", "docs/board/tasks/A1.md", taskFile("id: A1\ntitle: t", ""), "missing required field status"},
		{"filename", "docs/board/tasks/B2-x.md", taskFile("id: A1\ntitle: t\nstatus: backlog", ""), "filename must start with the id"},
		{"prefix is not enough", "docs/board/tasks/A10.md", taskFile("id: A1\ntitle: t\nstatus: backlog", ""), "filename must start with the id"},
		{"done outside archive", "docs/board/tasks/A1.md", taskFile("id: A1\ntitle: t\nstatus: done", ""), "only valid in the archive"},
		{"nested", "docs/board/tasks/old/A1.md", taskFile("id: A1\ntitle: t\nstatus: backlog", ""), "sub-folder"},
		{"title is a list", "docs/board/tasks/A1.md", taskFile("id: A1\ntitle: [a, b]\nstatus: backlog", ""), `field "title"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Parse(map[string][]byte{c.path: c.file}, layout)
			if len(r.Tasks) != 0 || len(r.Problems) != 1 {
				t.Fatalf("want 1 problem and no tasks, got %d tasks, problems %+v", len(r.Tasks), r.Problems)
			}
			if !strings.Contains(r.Problems[0].Reason, c.reason) {
				t.Fatalf("reason %q does not mention %q", r.Problems[0].Reason, c.reason)
			}
		})
	}
}

func TestParseKeepsGoodFilesNextToBadOnes(t *testing.T) {
	files := map[string][]byte{
		"docs/board/tasks/A1.md":      taskFile("id: A1\ntitle: one\nstatus: backlog", ""),
		"docs/board/tasks/A1-copy.md": taskFile("id: A1\ntitle: copy\nstatus: backlog", ""),
		"docs/board/tasks/A2.md":      []byte("not a task"),
		"docs/board/archive/A3.md":    taskFile("id: A3\ntitle: old\nstatus: done\ndone: 2026-09-01", ""),
		"docs/board/README.md":        []byte("# not a task"),
		"docs/other/A4.md":            taskFile("id: A4\ntitle: elsewhere\nstatus: backlog", ""),
	}
	r := Parse(files, layout)
	if len(r.Tasks) != 1 || r.Tasks[0].Path != "docs/board/tasks/A1-copy.md" {
		t.Fatalf("want only the first A1 by path (A1-copy.md sorts before A1.md), got %+v", r.Tasks)
	}
	if len(r.Archived) != 1 || !r.Archived[0].Archived || r.Archived[0].Done != "2026-09-01" {
		t.Fatalf("want archived A3, got %+v", r.Archived)
	}
	if len(r.Problems) != 2 {
		t.Fatalf("want duplicate + broken file problems, got %+v", r.Problems)
	}
}

func TestParseConfig(t *testing.T) {
	files := map[string][]byte{
		"docs/board/stages.md": []byte(`---
stages:
  - name: Backlog
    status: [backlog]
  - name: In review
    when: pr-open
  - name: Testing
    status: [testing, in qa]
---
# Stages
`),
		"docs/board/labels.md": []byte(`---
groups:
  type: [bug, feat]
  epic: []
colors:
  "type:bug": "#ff0000"
---
`),
		"docs/board/people.md": []byte(`---
people:
  - name: Nemanja Mudrinic
    aliases: [Nemanja]
---
| table | stays |
`),
		"docs/board/repos.md": []byte("---\ncode_repos: [acme/app, acme/web]\n---\n"),
	}
	r := Parse(files, layout)
	if len(r.Problems) != 0 {
		t.Fatalf("unexpected problems: %+v", r.Problems)
	}
	checkDeep(t, "stages", r.Config.Stages, []Stage{
		{Name: "Backlog", Statuses: []string{"backlog"}},
		{Name: "In review", WhenPROpen: true},
		{Name: "Testing", Statuses: []string{"testing", "in qa"}},
	})
	checkDeep(t, "label groups", r.Config.LabelGroups, []LabelGroup{
		{Name: "type", Values: []string{"bug", "feat"}},
		{Name: "epic", Values: []string{}},
	})
	checkDeep(t, "label colors", r.Config.LabelColors, map[string]string{"type:bug": "#ff0000"})
	checkDeep(t, "people", r.Config.People, []Person{{Name: "Nemanja Mudrinic", Aliases: []string{"Nemanja"}}})
	checkDeep(t, "code repos", r.Config.CodeRepos, []string{"acme/app", "acme/web"})
}

func TestParseInvalidConfigIsIgnored(t *testing.T) {
	cases := map[string]string{
		"docs/board/stages.md": "---\nstages:\n  - status: [backlog]\n---\n",
		"docs/board/labels.md": "---\ngroups: [type]\n---\n",
		"docs/board/people.md": "# no frontmatter\n",
		"docs/board/repos.md":  "---\ncode_repos: [not-a-repo]\n---\n",
	}
	for p, content := range cases {
		r := Parse(map[string][]byte{p: []byte(content)}, layout)
		if len(r.Problems) != 1 || r.Problems[0].Path != p {
			t.Fatalf("%s: want one problem, got %+v", p, r.Problems)
		}
		if r.Config.Stages != nil || r.Config.LabelGroups != nil || r.Config.People != nil || r.Config.CodeRepos != nil {
			t.Fatalf("%s: invalid config must be ignored, got %+v", p, r.Config)
		}
	}
}

func TestParseRealRepo(t *testing.T) {
	root := os.Getenv("FUDA_REAL_REPO")
	if root == "" {
		t.Skip("set FUDA_REAL_REPO to a checkout with docs/board to run")
	}
	files := map[string][]byte{}
	err := filepath.WalkDir(filepath.Join(root, "docs/board"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		content, err := os.ReadFile(p)
		files[filepath.ToSlash(rel)] = content
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	r := Parse(files, layout)

	counts := map[string]int{}
	for _, task := range r.Tasks {
		counts[task.Status]++
	}
	t.Logf("tasks %d, archived %d, by status %v", len(r.Tasks), len(r.Archived), counts)
	t.Logf("config: %d stages, %d label groups, %d people", len(r.Config.Stages), len(r.Config.LabelGroups), len(r.Config.People))
	for _, p := range r.Problems {
		t.Errorf("problem: %s: %s", p.Path, p.Reason)
	}
}

func check(t *testing.T, what, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", what, got, want)
	}
}

func checkDeep(t *testing.T, what string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s:\n got  %#v\n want %#v", what, got, want)
	}
}
