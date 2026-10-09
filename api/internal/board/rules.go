package board

import (
	"strings"

	"fuda/internal/taskfiles"
)

func Search(tasks []taskfiles.Task, query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	out := []string{}
	if query == "" {
		return out
	}
	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Body), query) || strings.Contains(strings.ToLower(t.Title), query) {
			out = append(out, t.ID)
		}
	}
	return out
}
