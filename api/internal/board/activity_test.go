package board

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"fuda/internal/taskfiles"
)

func activities(entries ...taskfiles.Activity) func(*taskfiles.Task) {
	return func(t *taskfiles.Task) { t.Activities = entries }
}

func TestCardShowsTheLatestActivityOnlyInAnActivityColumn(t *testing.T) {
	first := taskfiles.Activity{Text: "first"}
	latest := taskfiles.Activity{When: "2026-10-27 10:40", Actor: "builder", Event: "build", Result: "PASS", Text: "latest", Details: []string{"a finding"}}
	b := Build(Inputs{Develop: taskfiles.Result{
		Tasks: []taskfiles.Task{
			task("A1", "in progress", activities(first, latest)),
			task("A2", "backlog", activities(first)),
			task("A3", "in progress"),
		},
		Config: taskfiles.Config{Stages: []taskfiles.Stage{
			{Name: "Backlog", Statuses: []string{"backlog"}},
			{Name: "In progress", Statuses: []string{"in progress"}, Activity: true},
		}},
	}})

	onCard := latest
	onCard.Details = nil
	if got := card(b, "A1").Activity; got == nil || !reflect.DeepEqual(*got, onCard) {
		t.Errorf("A1 activity: got %+v, want the latest entry without its details", got)
	}
	if got := card(b, "A2").Activity; got != nil {
		t.Errorf("a column without activity: true shows no entry, got %+v", got)
	}
	if got := card(b, "A3").Activity; got != nil {
		t.Errorf("a task without Activity shows no entry, got %+v", got)
	}
	if !b.Columns[1].Activity || b.Columns[0].Activity {
		t.Errorf("column activity flags: %+v", b.Columns)
	}
}

func TestTaskViewSplitsTheActivityFromTheBody(t *testing.T) {
	files := map[string][]byte{"docs/board/tasks/A1.md": []byte("---\nid: A1\ntitle: t\nstatus: backlog\n---\n" +
		"## Done when\nit is.\n\n## Activity\n\n- 2026-10-27 09:12 · factory · spec · READY · 12k\n- a free note\n")}
	s := NewService(fakeSource{files}, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	view, err := s.Task("A1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(view.HTML, "Activity") || strings.Contains(view.HTML, "factory") {
		t.Errorf("html still has the Activity section: %s", view.HTML)
	}
	if !strings.Contains(view.HTML, "it is.") {
		t.Errorf("html lost the rest of the body: %s", view.HTML)
	}
	want := []taskfiles.Activity{
		{When: "2026-10-27 09:12", Actor: "factory", Event: "spec", Result: "READY", Usage: "12k", Text: "2026-10-27 09:12 · factory · spec · READY · 12k"},
		{Text: "a free note"},
	}
	if !reflect.DeepEqual(view.Activities, want) {
		t.Errorf("activities: got %+v, want %+v", view.Activities, want)
	}
}
