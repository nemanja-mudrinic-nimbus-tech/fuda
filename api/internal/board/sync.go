package board

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"fuda/internal/markdown"
	"fuda/internal/taskfiles"
)

const syncTimeout = 2 * time.Minute

func (s *Service) Sync(ctx context.Context) error {
	_, err, _ := s.group.Do("sync", func() (any, error) {
		ctx, cancel := context.WithTimeout(ctx, syncTimeout)
		defer cancel()
		return nil, s.sync(ctx)
	})
	return err
}

func (s *Service) Poll(ctx context.Context) error {
	head, err := s.source.Head(ctx, s.opts.WorkBranch)
	if errors.Is(err, ErrBranchMissing) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	snap := s.snapshot.Load()
	if snap == nil || snap.heads[s.opts.WorkBranch] != head {
		err := s.Sync(ctx)
		if errors.Is(err, ErrBranchMissing) {
			return ErrNotFound
		}
		return err
	}
	if s.refreshDue() {
		go func() { _ = s.Sync(context.WithoutCancel(ctx)) }()
	}
	return nil
}

func (s *Service) refreshDue() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.now().Sub(s.status.LastAttempt) >= s.opts.Cooldown
}

func (s *Service) RequestSync(ctx context.Context) bool {
	s.mu.Lock()
	now := s.now()
	if !s.lastRequest.IsZero() && now.Sub(s.lastRequest) < s.opts.Cooldown {
		s.mu.Unlock()
		return false
	}
	s.lastRequest = now
	s.mu.Unlock()

	go func() { _ = s.Sync(context.WithoutCancel(ctx)) }()
	return true
}

func (s *Service) sync(ctx context.Context) error {
	err := s.load(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastAttempt = s.now()
	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) {
		return err
	}
	if err != nil {
		s.status.LastError = err.Error()
		return err
	}
	s.status.LastError = ""
	return nil
}

func (s *Service) load(ctx context.Context) error {
	developHead, err := s.source.Head(ctx, s.opts.WorkBranch)
	if err != nil {
		return fmt.Errorf("%s: %w", s.opts.WorkBranch, err)
	}
	heads := map[string]string{s.opts.WorkBranch: developHead}
	if s.opts.WatchMain {
		mainHead, err := s.source.Head(ctx, s.opts.ProdBranch)
		switch {
		case errors.Is(err, ErrBranchMissing):
		case err != nil:
			return fmt.Errorf("%s: %w", s.opts.ProdBranch, err)
		default:
			heads[s.opts.ProdBranch] = mainHead
		}
	}

	current := s.snapshot.Load()
	if current != nil && maps.Equal(current.heads, heads) {
		found, err := s.openReviews(ctx, current.develop)
		if err != nil {
			s.markSynced(heads, current)
			return fmt.Errorf("open pull requests: %w", err)
		}
		if !sameReviews(current.reviews, found) {
			current = s.build(current.develop, current.main, current.files, heads, found)
			s.snapshot.Store(current)
		}
		s.markSynced(heads, current)
		return nil
	}

	developFiles, err := s.source.Files(ctx, s.opts.WorkBranch)
	if err != nil {
		return fmt.Errorf("%s: %w", s.opts.WorkBranch, err)
	}
	develop := taskfiles.Parse(developFiles, s.layout())

	var mainFiles map[string][]byte
	if _, ok := heads[s.opts.ProdBranch]; ok && s.opts.WatchMain {
		mainFiles, err = s.source.Files(ctx, s.opts.ProdBranch)
		if err != nil {
			return fmt.Errorf("%s: %w", s.opts.ProdBranch, err)
		}
	}
	if err := s.saveCache(heads, developFiles, mainFiles); err != nil {
		s.opts.Logger.Warn("the disk cache was not updated", "error", err)
	}
	main := s.parseMain(mainFiles)

	found, reviewErr := s.openReviews(ctx, develop)
	if reviewErr != nil {
		found = reviews{}
	}
	next := s.build(develop, main, developFiles, heads, found)
	s.snapshot.Store(next)
	s.markSynced(heads, next)
	if reviewErr != nil {
		return fmt.Errorf("open pull requests: %w", reviewErr)
	}
	return nil
}

func (s *Service) layout() taskfiles.Layout {
	return taskfiles.Layout{BoardDir: s.opts.BoardDir}
}

func (s *Service) parseMain(files map[string][]byte) *taskfiles.Result {
	if files == nil {
		return nil
	}
	parsed := taskfiles.Parse(files, s.layout())
	if len(parsed.Tasks) == 0 && len(parsed.Archived) == 0 {
		return nil
	}
	return &parsed
}

func (s *Service) build(develop taskfiles.Result, main *taskfiles.Result, files map[string][]byte, heads map[string]string, found reviews) *snapshot {
	docs := map[string][]byte{}
	docSet := map[string]bool{}
	assets := map[string][]byte{}
	assetSet := map[string]bool{}
	for p, content := range files {
		if !strings.HasPrefix(p, s.opts.DocsRoot+"/") {
			continue
		}
		if strings.HasSuffix(p, ".md") {
			docs[p] = content
			docSet[p] = true
		} else if _, ok := AssetType(p); ok {
			assets[p] = content
			assetSet[p] = true
		}
	}
	byID := map[string]taskfiles.Task{}
	taskPaths := map[string]string{}
	for _, t := range slices.Concat(develop.Tasks, develop.Archived) {
		byID[t.ID] = t
		taskPaths[t.Path] = t.ID
	}

	return &snapshot{
		board: Build(Inputs{
			Develop: develop,
			Main:    main,
			OpenPRs: found,
		}),
		develop: develop,
		main:    main,
		files:   files,
		reviews: found,
		docs:    docs,
		assets:  assets,
		byID:    byID,
		heads:   heads,
		links: markdown.Links{
			Prefix:    s.opts.Origin.Path,
			DocsRoot:  s.opts.DocsRoot,
			Docs:      docSet,
			Assets:    assetSet,
			TaskPaths: taskPaths,
			CodeURL:   s.opts.CodeURL,
		},
	}
}

func (s *Service) markSynced(heads map[string]string, snap *snapshot) {
	s.markSyncedAt(heads, snap, s.now())
}

func (s *Service) markSyncedAt(heads map[string]string, snap *snapshot, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Develop = BranchStatus{Branch: s.opts.WorkBranch, SHA: heads[s.opts.WorkBranch], SyncedAt: now}
	if !s.opts.WatchMain {
		s.status.Main = nil
		return
	}
	s.status.Main = &BranchStatus{Branch: s.opts.ProdBranch, SHA: heads[s.opts.ProdBranch], SyncedAt: now, NotYet: snap.main == nil}
}
