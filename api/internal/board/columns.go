package board

import (
	"strings"

	"fuda/internal/taskfiles"
)

var defaultStages = []taskfiles.Stage{
	{Name: "Backlog", Statuses: []string{"backlog"}},
	{Name: "In progress", Statuses: []string{"in progress"}},
	{Name: "In review", WhenPROpen: true},
	{Name: "Merged", Statuses: []string{"merged"}},
	{Name: "Testing", Statuses: []string{"testing"}},
	{Name: "Validated", Statuses: []string{"validated"}},
}

type columns struct {
	list     []Column
	byStatus map[string]string
	review   string
	done     *Column
	activity map[string]bool
}

func buildColumns(stages []taskfiles.Stage, tasks []taskfiles.Task) *columns {
	if stages == nil {
		stages = defaultStages
	}
	c := &columns{byStatus: map[string]string{}, activity: map[string]bool{}}
	for _, s := range stages {
		col := Column{ID: slug(s.Name), Name: s.Name, Statuses: orEmpty(s.Statuses), PROpen: s.WhenPROpen, Done: s.WhenDone, Days: s.Days, Activity: s.Activity}
		if s.Activity {
			c.activity[col.ID] = true
		}
		if s.WhenDone {
			c.done = &col
			continue
		}
		c.list = append(c.list, col)
		for _, status := range s.Statuses {
			c.byStatus[normalizeStatus(status)] = col.ID
		}
		if s.WhenPROpen && c.review == "" {
			c.review = col.ID
		}
	}
	for _, t := range tasks {
		status := normalizeStatus(t.Status)
		if _, known := c.byStatus[status]; known {
			continue
		}
		col := Column{ID: "unknown-" + slug(status), Name: t.Status, Statuses: []string{t.Status}, Unknown: true}
		c.list = append(c.list, col)
		c.byStatus[status] = col.ID
	}
	if c.done != nil {
		c.list = append(c.list, *c.done)
	}
	return c
}

func (c *columns) place(status string, prOpen bool) string {
	if prOpen && c.review != "" {
		return c.review
	}
	return c.byStatus[normalizeStatus(status)]
}

func (c *columns) latestActivity(column string, entries []taskfiles.Activity) *taskfiles.Activity {
	if !c.activity[column] || len(entries) == 0 {
		return nil
	}
	latest := entries[len(entries)-1]
	return &latest
}

func normalizeStatus(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
