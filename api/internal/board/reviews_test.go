package board

import (
	"context"
	"errors"
	"testing"
)

type reviewHost struct {
	*fakeHost
	prs  map[string][]PullRequest
	errs map[string]error
}

func (h *reviewHost) OpenPRs(_ context.Context, repo string) ([]PullRequest, error) {
	return h.prs[repo], h.errs[repo]
}

func newReviewService(t *testing.T, prs map[string][]PullRequest, errs map[string]error) (*Service, *reviewHost) {
	t.Helper()
	files := map[string]string{
		"docs/board/repos.md":      "---\ncode_repos: [acme/app, acme/web]\n---\n",
		"docs/board/tasks/T-1.md":  taskContent("in progress", ""),
		"docs/board/tasks/T-12.md": "---\nid: T-12\ntitle: Twelve\nstatus: backlog\n---\n",
		"docs/board/tasks/T-2.md":  "---\nid: T-2\ntitle: Two\nstatus: backlog\n---\n",
	}
	host := &reviewHost{fakeHost: newFakeHost(files), prs: prs, errs: errs}
	s := NewService(host, Options{DocsRoot: "docs", BoardDir: "docs/board", WorkBranch: "develop", ProdBranch: "main"})
	err := s.Poll(context.Background())
	if _, ready := s.Board(); !ready {
		t.Fatal(err)
	}
	return s, host
}

func TestOpenPRPutsTasksNamedInTitleOrBranchInReview(t *testing.T) {
	s, _ := newReviewService(t, map[string][]PullRequest{
		"acme/app": {
			{Number: 5, URL: "u5", Title: "T-1: add the thing", Branch: "feature/x"},
			{Number: 6, Title: "Fix a bug", Branch: "bugfix/t-2-login"},
		},
	}, nil)

	view, _ := s.Board()
	if got := card(view.Board, "T-1"); got.Column != "in-review" || len(got.OpenPRs) != 1 || got.OpenPRs[0] != (OpenPR{Number: 5, URL: "u5"}) {
		t.Errorf("T-1 named in a title: %+v", got)
	}
	if got := card(view.Board, "T-2").Column; got != "in-review" {
		t.Errorf("T-2 named in a branch: %q", got)
	}
	if got := card(view.Board, "T-12").Column; got != "backlog" {
		t.Errorf("T-12 is not named by T-1 or T-2: %q", got)
	}
}

func TestOnePRNamingTwoTasksPutsBothInReview(t *testing.T) {
	s, _ := newReviewService(t, map[string][]PullRequest{
		"acme/web": {{Number: 9, Title: "T-1 and T-12 together", Branch: "main-work"}},
	}, nil)

	for _, id := range []string{"T-1", "T-12"} {
		if got := cardColumn(t, s, id); got != "in-review" {
			t.Errorf("%s: %q", id, got)
		}
	}
	if got := cardColumn(t, s, "T-2"); got != "backlog" {
		t.Errorf("T-2: %q", got)
	}
}

func TestClosedPRReturnsTheCardToItsStatusStage(t *testing.T) {
	s, host := newReviewService(t, map[string][]PullRequest{
		"acme/app": {{Number: 5, Title: "T-1 work"}},
	}, nil)
	if got := cardColumn(t, s, "T-1"); got != "in-review" {
		t.Fatalf("with the PR open: %q", got)
	}

	host.prs = nil
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := cardColumn(t, s, "T-1"); got != "in-progress" {
		t.Errorf("after the PR closed: %q", got)
	}
}

func TestLockedCardRefusesMove(t *testing.T) {
	s, host := newReviewService(t, map[string][]PullRequest{
		"acme/app": {{Number: 5, Title: "T-1 work"}},
	}, nil)

	err := s.Move(context.Background(), MoveRequest{TaskID: "T-1", Seen: "in progress", Column: "backlog"})
	if !errors.Is(err, ErrLocked) {
		t.Fatalf("got %v", err)
	}
	if host.writes != 0 {
		t.Error("a refused Move must not write")
	}
}

func TestRepositoryTheUserCannotSeeIsSkipped(t *testing.T) {
	for _, hidden := range []error{ErrNotFound, ErrForbidden} {
		s, _ := newReviewService(t, map[string][]PullRequest{
			"acme/web": {{Number: 9, Title: "T-2 work"}},
		}, map[string]error{"acme/app": hidden})

		if _, ready := s.Board(); !ready {
			t.Fatalf("%v: the Board must load", hidden)
		}
		if got := cardColumn(t, s, "T-2"); got != "in-review" {
			t.Errorf("%v: the readable repository still counts: %q", hidden, got)
		}
		if got := s.Status().LastError; got != "" {
			t.Errorf("%v: a hidden repository is not an error: %q", hidden, got)
		}
	}
}

func TestFailingRepositoryKeepsTheBoardAndReportsTheError(t *testing.T) {
	s, _ := newReviewService(t, nil, map[string]error{"acme/app": errors.New("boom")})

	if got := cardColumn(t, s, "T-1"); got != "in-progress" {
		t.Errorf("T-1: %q", got)
	}
	if s.Status().LastError == "" {
		t.Error("the failure must be reported")
	}
}
