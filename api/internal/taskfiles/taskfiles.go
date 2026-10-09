package taskfiles

import (
	"path"
	"slices"
	"strings"
)

type Task struct {
	Path      string
	ID        string
	Title     string
	Status    string
	BlockedBy string
	Owners    []string
	Testers   []string
	Labels    []Label
	Added     string
	Claimed   string
	Done      string
	PRs       []int
	PRRef     string
	Custom    []Field
	Body      string
	Archived  bool
}

type Label struct {
	Group string
	Value string
}

func (l Label) String() string {
	if l.Group == "" {
		return l.Value
	}
	return l.Group + ":" + l.Value
}

type Field struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Stage struct {
	Name       string
	Statuses   []string
	WhenPROpen bool
}

type LabelGroup struct {
	Name   string
	Values []string
}

type Person struct {
	Name    string
	Aliases []string
}

type Config struct {
	Stages      []Stage
	LabelGroups []LabelGroup
	LabelColors map[string]string
	People      []Person
	CodeRepos   []string
}

type Problem struct {
	Path   string
	Reason string
}

type Result struct {
	Tasks    []Task
	Archived []Task
	Config   Config
	Problems []Problem
}

type Layout struct {
	BoardDir string
}

func (l Layout) tasksDir() string   { return path.Join(l.BoardDir, "tasks") }
func (l Layout) archiveDir() string { return path.Join(l.BoardDir, "archive") }

func Parse(files map[string][]byte, layout Layout) Result {
	var r Result
	r.Config = parseConfig(files, layout, &r.Problems)

	seen := map[string]string{}
	for _, p := range sortedPaths(files) {
		archived, ok := taskLocation(p, layout)
		if !ok {
			if nestedUnder(p, layout.tasksDir()) || nestedUnder(p, layout.archiveDir()) {
				r.Problems = append(r.Problems, Problem{Path: p, Reason: "task files must sit directly in tasks/ or archive/, not in a sub-folder"})
			}
			continue
		}
		task, reason := parseTask(p, files[p], archived)
		if reason == "" {
			if first, dup := seen[task.ID]; dup {
				reason = "duplicate id " + task.ID + " (also in " + first + ")"
			}
		}
		if reason != "" {
			r.Problems = append(r.Problems, Problem{Path: p, Reason: reason})
			continue
		}
		seen[task.ID] = p
		if archived {
			r.Archived = append(r.Archived, task)
		} else {
			r.Tasks = append(r.Tasks, task)
		}
	}
	return r
}

func taskLocation(p string, layout Layout) (archived, ok bool) {
	if !strings.HasSuffix(p, ".md") {
		return false, false
	}
	switch path.Dir(p) {
	case layout.tasksDir():
		return false, true
	case layout.archiveDir():
		return true, true
	}
	return false, false
}

func nestedUnder(p, dir string) bool {
	return strings.HasSuffix(p, ".md") && strings.HasPrefix(p, dir+"/") && path.Dir(p) != dir
}

func sortedPaths(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	slices.Sort(paths)
	return paths
}
