package taskfiles

import (
	"os"
	"testing"
)

func TestSplitActivity(t *testing.T) {
	body, err := os.ReadFile("testdata/activity.md")
	if err != nil {
		t.Fatal(err)
	}
	rest, entries := SplitActivity(string(body))

	check(t, "rest", rest, "# A1\n\n## Done when\nit is.\n\n## Notes\nstays in the body.\n")
	checkDeep(t, "entries", entries, []Activity{
		{When: "2026-10-27 09:12", Actor: "factory", Event: "spec", Result: "READY", Usage: "12k tokens", Text: "2026-10-27 09:12 · factory · spec · READY · 12k tokens"},
		{When: "2026-10-27 10:40", Actor: "builder", Event: "build", Result: "PASS", Text: "2026-10-27 10:40 · builder · build · PASS"},
		{Text: "2026-10-27 11:00 · reviewer · code review · CHANGES · 3 findings · extra"},
		{Text: "the reviewer asked for a second look"},
		{Text: "2026-10-27 · nobody · no time · PASS"},
		{Text: "```"},
		{Text: "## not a heading"},
		{Text: "```"},
	})
}

func TestSplitActivityWithoutTheSection(t *testing.T) {
	body := "# A1\n\n## Done when\nit is.\n"
	rest, entries := SplitActivity(body)
	check(t, "rest", rest, body)
	if entries != nil {
		t.Errorf("want no entries, got %+v", entries)
	}
}

func TestActivityAtTheEndAndEmpty(t *testing.T) {
	rest, entries := SplitActivity("intro\n## Activity\n")
	check(t, "rest", rest, "intro")
	if len(entries) != 0 {
		t.Errorf("want no entries, got %+v", entries)
	}
}

func TestTaskWithBrokenActivityIsNoProblem(t *testing.T) {
	files := map[string][]byte{
		"docs/board/tasks/A1.md": taskFile("id: A1\ntitle: t\nstatus: backlog", "## Activity\n- ·· ·\n- 2026-10-27 09:12 · factory · spec · READY\n"),
	}
	r := Parse(files, layout)
	if len(r.Problems) != 0 || len(r.Tasks) != 1 {
		t.Fatalf("want 1 task and no problems, got %d tasks, problems %+v", len(r.Tasks), r.Problems)
	}
	checkDeep(t, "activities", r.Tasks[0].Activities, []Activity{
		{Text: "·· ·"},
		{When: "2026-10-27 09:12", Actor: "factory", Event: "spec", Result: "READY", Text: "2026-10-27 09:12 · factory · spec · READY"},
	})
}
