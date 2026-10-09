package board

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"

	"fuda/internal/taskfiles"
)

type PullRequest struct {
	Number int
	URL    string
	Title  string
	Branch string
}

type ReviewSource interface {
	OpenPRs(ctx context.Context, repo string) ([]PullRequest, error)
}

type OpenPR struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type reviews map[string][]OpenPR

func (s *Service) openReviews(ctx context.Context, develop taskfiles.Result) (reviews, error) {
	found := reviews{}
	source, ok := s.source.(ReviewSource)
	if !ok {
		return found, nil
	}
	for _, repo := range develop.Config.CodeRepos {
		pulls, err := source.OpenPRs(ctx, repo)
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrForbidden) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, pr := range pulls {
			for _, id := range taskIDsIn(develop.Tasks, pr.Title, pr.Branch) {
				found[id] = append(found[id], OpenPR{Number: pr.Number, URL: pr.URL})
			}
		}
	}
	return found, nil
}

func taskIDsIn(tasks []taskfiles.Task, texts ...string) []string {
	haystack := strings.ToLower(strings.Join(texts, "\n"))
	var ids []string
	for _, t := range tasks {
		if containsID(haystack, strings.ToLower(t.ID)) {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

func containsID(haystack, id string) bool {
	for offset := 0; ; {
		i := strings.Index(haystack[offset:], id)
		if i < 0 {
			return false
		}
		start, end := offset+i, offset+i+len(id)
		if !isIDChar(haystack, start-1) && !isIDChar(haystack, end) {
			return true
		}
		offset = start + 1
	}
}

func isIDChar(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

func sameReviews(a, b reviews) bool {
	return maps.EqualFunc(a, b, slices.Equal[[]OpenPR])
}
