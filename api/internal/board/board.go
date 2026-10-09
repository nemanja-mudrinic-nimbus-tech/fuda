package board

import (
	"fuda/internal/taskfiles"
)

type Board struct {
	Columns  []Column  `json:"columns"`
	Cards    []Card    `json:"cards"`
	Facets   Facets    `json:"facets"`
	Problems []Problem `json:"problems"`
}

type Column struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Statuses []string `json:"statuses"`
	PROpen   bool     `json:"prOpen"`
	Unknown  bool     `json:"unknown"`
}

type Card struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Column       string   `json:"column"`
	Owners       []string `json:"owners"`
	Testers      []string `json:"testers"`
	Labels       []string `json:"labels"`
	Prefix       string   `json:"prefix"`
	Added        string   `json:"added"`
	Claimed      string   `json:"claimed"`
	Done         string   `json:"done,omitempty"`
	PRs          []int    `json:"prs"`
	PRRef        string   `json:"prRef,omitempty"`
	OpenPRs      []OpenPR `json:"openPrs"`
	BlockedBy    string   `json:"blockedBy,omitempty"`
	BlockedByIDs []string `json:"blockedByIds"`
	Blocks       []string `json:"blocks"`
	References   []string `json:"references"`
	ReferencedBy []string `json:"referencedBy"`
	InProd       bool     `json:"inProd"`
	Path         string   `json:"path"`
}

type Facets struct {
	Statuses    []string          `json:"statuses"`
	People      []string          `json:"people"`
	Testers     []string          `json:"testers"`
	LabelGroups []LabelGroup      `json:"labelGroups"`
	LabelColors map[string]string `json:"labelColors"`
	Prefixes    []string          `json:"prefixes"`
}

type LabelGroup struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type Problem struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type Inputs struct {
	Develop taskfiles.Result
	Main    *taskfiles.Result
	OpenPRs map[string][]OpenPR
}

var inProdStatuses = map[string]bool{"merged": true, "testing": true, "validated": true}

func Build(in Inputs) Board {
	tasks := in.Develop.Tasks
	people := newPeople(in.Develop.Config.People, tasks)
	refs := newReferences(tasks)
	columns := buildColumns(in.Develop.Config.Stages, tasks)
	prodStatus := mainStatuses(in.Main)

	cards := make([]Card, 0, len(tasks))
	for _, t := range tasks {
		openPRs := in.OpenPRs[t.ID]
		inProd := inProdStatuses[normalizeStatus(prodStatus[t.ID])]
		cards = append(cards, Card{
			ID:           t.ID,
			Title:        t.Title,
			Status:       t.Status,
			Column:       columns.place(t.Status, len(openPRs) > 0),
			Owners:       people.canonical(t.Owners),
			Testers:      people.canonical(t.Testers),
			Labels:       labelStrings(t.Labels),
			Prefix:       idPrefix(t.ID),
			Added:        t.Added,
			Claimed:      t.Claimed,
			PRs:          orEmpty(t.PRs),
			PRRef:        t.PRRef,
			OpenPRs:      orEmpty(openPRs),
			BlockedBy:    t.BlockedBy,
			BlockedByIDs: refs.blockedBy[t.ID],
			Blocks:       refs.blocks[t.ID],
			References:   refs.mentions[t.ID],
			ReferencedBy: refs.mentionedBy[t.ID],
			InProd:       inProd,
			Path:         t.Path,
		})
	}

	return Board{
		Columns:  columns.list,
		Cards:    cards,
		Facets:   buildFacets(in.Develop, people),
		Problems: problems(in.Develop.Problems),
	}
}

func mainStatuses(main *taskfiles.Result) map[string]string {
	statuses := map[string]string{}
	if main == nil {
		return statuses
	}
	for _, t := range main.Tasks {
		statuses[t.ID] = t.Status
	}
	return statuses
}

func problems(in []taskfiles.Problem) []Problem {
	out := make([]Problem, 0, len(in))
	for _, p := range in {
		out = append(out, Problem(p))
	}
	return out
}

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
