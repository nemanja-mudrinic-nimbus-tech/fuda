package board

import (
	"cmp"
	"slices"
	"strings"

	"fuda/internal/taskfiles"
)

type references struct {
	blockedBy   map[string][]string
	blocks      map[string][]string
	mentions    map[string][]string
	mentionedBy map[string][]string
}

func newReferences(tasks []taskfiles.Task) references {
	known := map[string]bool{}
	for _, t := range tasks {
		known[t.ID] = true
	}
	r := references{
		blockedBy:   map[string][]string{},
		blocks:      map[string][]string{},
		mentions:    map[string][]string{},
		mentionedBy: map[string][]string{},
	}
	for _, t := range tasks {
		for _, id := range KnownIDs(t.BlockedBy, known) {
			if id == t.ID {
				continue
			}
			r.blockedBy[t.ID] = append(r.blockedBy[t.ID], id)
			r.blocks[id] = append(r.blocks[id], t.ID)
		}
		for _, id := range KnownIDs(t.Body, known) {
			if id == t.ID {
				continue
			}
			r.mentions[t.ID] = append(r.mentions[t.ID], id)
			r.mentionedBy[id] = append(r.mentionedBy[id], t.ID)
		}
	}
	for _, m := range []map[string][]string{r.blockedBy, r.blocks, r.mentions, r.mentionedBy} {
		for _, t := range tasks {
			m[t.ID] = orEmpty(m[t.ID])
		}
	}
	return r
}

func KnownIDs(text string, known map[string]bool) []string {
	haystack := strings.ToLower(text)
	type found struct {
		id string
		at int
	}
	var hits []found
	for id := range known {
		if at := indexOfID(haystack, strings.ToLower(id)); at >= 0 {
			hits = append(hits, found{id, at})
		}
	}
	slices.SortFunc(hits, func(a, b found) int { return cmp.Or(cmp.Compare(a.at, b.at), cmp.Compare(a.id, b.id)) })
	out := make([]string, len(hits))
	for i, hit := range hits {
		out[i] = hit.id
	}
	return out
}

func indexOfID(haystack, id string) int {
	for offset := 0; offset < len(haystack); {
		i := strings.Index(haystack[offset:], id)
		if i < 0 {
			return -1
		}
		start, end := offset+i, offset+i+len(id)
		if !isIDChar(haystack, start-1) && !isIDChar(haystack, end) {
			return start
		}
		offset = start + 1
	}
	return -1
}

func isIDChar(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}
