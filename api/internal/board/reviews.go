package board

import (
	"context"
	"errors"
	"maps"
	"slices"

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

type LinkedPR struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type prsByTask map[string][]LinkedPR

func (s *Service) openReviews(ctx context.Context, develop taskfiles.Result) (prsByTask, error) {
	found := prsByTask{}
	source, ok := s.source.(ReviewSource)
	if !ok {
		return found, nil
	}
	known := map[string]bool{}
	for _, t := range develop.Tasks {
		known[t.ID] = true
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
			for _, id := range KnownIDs(pr.Title+"\n"+pr.Branch, known) {
				found[id] = append(found[id], LinkedPR{Number: pr.Number, URL: pr.URL})
			}
		}
	}
	return found, nil
}

func sameReviews(a, b prsByTask) bool {
	return maps.EqualFunc(a, b, slices.Equal[[]LinkedPR])
}
