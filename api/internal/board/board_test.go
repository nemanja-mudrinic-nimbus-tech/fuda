package board

import (
	"reflect"
	"testing"

	"fuda/internal/taskfiles"
)

func task(id, status string, mods ...func(*taskfiles.Task)) taskfiles.Task {
	t := taskfiles.Task{ID: id, Title: id + " title", Status: status, Path: "docs/board/tasks/" + id + ".md"}
	for _, m := range mods {
		m(&t)
	}
	return t
}

func owners(names ...string) func(*taskfiles.Task) {
	return func(t *taskfiles.Task) { t.Owners = names }
}

func labels(ls ...taskfiles.Label) func(*taskfiles.Task) {
	return func(t *taskfiles.Task) { t.Labels = ls }
}

func body(b string) func(*taskfiles.Task) {
	return func(t *taskfiles.Task) { t.Body = b }
}

func blockedBy(b string) func(*taskfiles.Task) {
	return func(t *taskfiles.Task) { t.BlockedBy = b }
}

func columnOf(b Board, id string) string {
	for _, c := range b.Cards {
		if c.ID == id {
			return c.Column
		}
	}
	return ""
}

func card(b Board, id string) Card {
	for _, c := range b.Cards {
		if c.ID == id {
			return c
		}
	}
	return Card{}
}

func columnIDs(b Board) []string {
	ids := []string{}
	for _, c := range b.Columns {
		ids = append(ids, c.ID)
	}
	return ids
}

func TestDefaultColumnsWithoutStages(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{Tasks: []taskfiles.Task{
		task("A1", "backlog"),
		task("A2", "In Progress"),
		task("A3", "qa"),
	}}})

	want := []string{"backlog", "in-progress", "in-review", "merged", "testing", "validated", "unknown-qa"}
	if got := columnIDs(b); !reflect.DeepEqual(got, want) {
		t.Fatalf("columns: got %v, want %v", got, want)
	}
	if !b.Columns[len(b.Columns)-1].Unknown {
		t.Error("unlisted status must give a column marked unknown")
	}
	if got := columnOf(b, "A2"); got != "in-progress" {
		t.Errorf("status matching is case-insensitive: got %q", got)
	}
	if got := columnOf(b, "A1"); got != "backlog" {
		t.Errorf("A1 column %q", got)
	}
}

func TestConfiguredStagesWithSeveralStatuses(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{
		Tasks: []taskfiles.Task{task("A1", "in qa"), task("A2", "testing"), task("A3", "blocked-ish")},
		Config: taskfiles.Config{Stages: []taskfiles.Stage{
			{Name: "Todo", Statuses: []string{"backlog"}},
			{Name: "Testing", Statuses: []string{"testing", "in qa"}},
		}},
	}})

	if got := columnIDs(b); !reflect.DeepEqual(got, []string{"todo", "testing", "unknown-blocked-ish"}) {
		t.Fatalf("columns: %v", got)
	}
	if columnOf(b, "A1") != "testing" || columnOf(b, "A2") != "testing" {
		t.Error("both statuses of a stage go to its column")
	}
}

func TestOpenPRMovesToReviewOnlyWhenAStageAllowsIt(t *testing.T) {
	tasks := []taskfiles.Task{task("A1", "in progress")}
	prs := map[string][]OpenPR{"A1": {{Number: 42}}}

	withReview := Build(Inputs{Develop: taskfiles.Result{Tasks: tasks}, OpenPRs: prs})
	if columnOf(withReview, "A1") != "in-review" {
		t.Error("default stages include In review")
	}
	if got := card(withReview, "A1").OpenPRs; !reflect.DeepEqual(got, []OpenPR{{Number: 42}}) {
		t.Errorf("open PRs %v", got)
	}

	without := Build(Inputs{Develop: taskfiles.Result{Tasks: tasks, Config: taskfiles.Config{Stages: []taskfiles.Stage{
		{Name: "Doing", Statuses: []string{"in progress"}},
	}}}, OpenPRs: prs})
	if columnOf(without, "A1") != "doing" {
		t.Error("without a pr-open stage the status decides")
	}
}

func TestClaimedTaskStaysInItsStatus(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{Tasks: []taskfiles.Task{task("A1", "backlog", owners("Ana"))}}})
	if columnOf(b, "A1") != "backlog" || len(b.Problems) != 0 {
		t.Error("an owner does not move a task or raise a problem")
	}
}

func TestInProd(t *testing.T) {
	main := taskfiles.Result{Tasks: []taskfiles.Task{task("A1", "validated"), task("A2", "merged"), task("A3", "in progress")}}
	b := Build(Inputs{
		Develop: taskfiles.Result{Tasks: []taskfiles.Task{task("A1", "validated"), task("A2", "merged"), task("A3", "in progress")}},
		Main:    &main,
	})
	for _, id := range []string{"A1", "A2"} {
		if !card(b, id).InProd {
			t.Errorf("%s is merged or later on main", id)
		}
	}
	if c := card(b, "A3"); c.InProd {
		t.Error("in progress on main is not in prod")
	}

	unwatched := Build(Inputs{Develop: taskfiles.Result{Tasks: []taskfiles.Task{task("A1", "validated")}}})
	if card(unwatched, "A1").InProd {
		t.Error("without main nothing is in prod")
	}
}

func TestReferencesAndBlocks(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{Tasks: []taskfiles.Task{
		task("SS-1", "backlog", blockedBy("SS-2, waiting on data"), body("See SS-3. Not SS-30 or SS-1.")),
		task("SS-2", "backlog"),
		task("SS-3", "backlog"),
	}}})

	c := card(b, "SS-1")
	check(t, "blocked by", c.BlockedByIDs, []string{"SS-2"})
	check(t, "references", c.References, []string{"SS-3"})
	check(t, "blocks", card(b, "SS-2").Blocks, []string{"SS-1"})
	check(t, "referenced by", card(b, "SS-3").ReferencedBy, []string{"SS-1"})
}

func TestPeopleFromConfigAreStrict(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{
		Tasks: []taskfiles.Task{task("A1", "backlog", owners("nemanja", "Guest"))},
		Config: taskfiles.Config{People: []taskfiles.Person{
			{Name: "Nemanja Mudrinic", Aliases: []string{"Nemanja"}},
		}},
	}})
	check(t, "facet", b.Facets.People, []string{"Nemanja Mudrinic"})
	check(t, "card owners", card(b, "A1").Owners, []string{"Nemanja Mudrinic", "Guest"})
}

func TestPeopleDerivedFoldUniqueFirstNames(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{Tasks: []taskfiles.Task{
		task("A1", "backlog", owners("Nemanja", "Jack")),
		task("A2", "backlog", owners("Nemanja Mudrinic", "Jack Westbrock", "Jack Black")),
	}}})
	check(t, "facet", b.Facets.People, []string{"Jack", "Jack Black", "Jack Westbrock", "Nemanja Mudrinic"})
	check(t, "A1 owners", card(b, "A1").Owners, []string{"Nemanja Mudrinic", "Jack"})
}

func TestLabelGroups(t *testing.T) {
	tasks := []taskfiles.Task{task("A1", "backlog", labels(
		taskfiles.Label{Group: "theme", Value: "security"},
		taskfiles.Label{Group: "area", Value: "be"},
		taskfiles.Label{Group: "type", Value: "spike"},
		taskfiles.Label{Value: "loose"},
	))}

	derived := Build(Inputs{Develop: taskfiles.Result{Tasks: tasks}})
	check(t, "derived groups", derived.Facets.LabelGroups, []LabelGroup{
		{Name: "epic", Values: []string{"security"}},
		{Name: "area", Values: []string{"be"}},
		{Name: "type", Values: append(append([]string{}, defaultTypes...), "spike")},
		{Name: "label", Values: []string{"loose"}},
	})
	check(t, "card labels", card(derived, "A1").Labels, []string{"epic:security", "area:be", "type:spike", "label:loose"})

	configured := Build(Inputs{Develop: taskfiles.Result{Tasks: tasks, Config: taskfiles.Config{LabelGroups: []taskfiles.LabelGroup{
		{Name: "type", Values: []string{"bug"}},
		{Name: "theme", Values: []string{"security"}},
	}}}})
	check(t, "configured groups are strict", configured.Facets.LabelGroups, []LabelGroup{
		{Name: "type", Values: []string{"bug"}},
		{Name: "epic", Values: []string{"security"}},
	})
}

func TestPrefixes(t *testing.T) {
	for id, want := range map[string]string{"SS-12": "SS", "S3b": "S", "TG-1": "TG", "0.1": "0", "N": "N"} {
		if got := idPrefix(id); got != want {
			t.Errorf("idPrefix(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestSearch(t *testing.T) {
	tasks := []taskfiles.Task{task("A1", "backlog", body("Rotate the API key")), task("A2", "backlog")}
	check(t, "body match", Search(tasks, "api KEY"), []string{"A1"})
	check(t, "title match", Search(tasks, "a2 title"), []string{"A2"})
	check(t, "empty", Search(tasks, "  "), []string{})
}

func check(t *testing.T, what string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s:\n got  %#v\n want %#v", what, got, want)
	}
}

func TestLabelColorsAreDistinctWithinAGroupAndConfigurable(t *testing.T) {
	b := Build(Inputs{Develop: taskfiles.Result{
		Tasks: []taskfiles.Task{task("A1", "backlog", labels(
			taskfiles.Label{Group: "epic", Value: "one"},
			taskfiles.Label{Group: "epic", Value: "two"},
			taskfiles.Label{Group: "type", Value: "bug"},
			taskfiles.Label{Group: "theme", Value: "three"},
		))},
		Config: taskfiles.Config{LabelColors: map[string]string{"theme:three": "#123456"}},
	}})
	c := b.Facets.LabelColors
	if c["epic:one"] == c["epic:two"] {
		t.Errorf("values in one group share a color: %v", c)
	}
	if c["type:bug"] != "#ef4444" {
		t.Errorf("type bug keeps its default color, got %q", c["type:bug"])
	}
	if c["epic:three"] != "#123456" {
		t.Errorf("configured color (theme folds into epic) wins, got %q", c["epic:three"])
	}
}
