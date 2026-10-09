package board

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"

	"fuda/internal/markdown"
	"fuda/internal/taskfiles"
)

type Source interface {
	Head(ctx context.Context, branch string) (string, error)
	Files(ctx context.Context, branch string) (map[string][]byte, error)
}

var ErrBranchMissing = errors.New("branch not found")

var ErrNotFound = errors.New("not found")

var ErrUnauthorized = errors.New("login required")

var ErrForbidden = errors.New("no access")

type Options struct {
	Title      string
	DocsRoot   string
	BoardDir   string
	WorkBranch string
	ProdBranch string
	WatchMain  bool
	Cooldown   time.Duration
	CacheDir   string
	Logger     *slog.Logger
	CodeURL    func(repoPath string) string
	PRLink     string
	Origin     Origin
}

type Origin struct {
	Host string `json:"host"`
	Repo string `json:"repo"`
	URL  string `json:"url,omitempty"`
	Path string `json:"path"`
}

type Service struct {
	source Source
	opts   Options
	now    func() time.Time

	snapshot atomic.Pointer[snapshot]
	group    singleflight.Group

	mu          sync.Mutex
	lastRequest time.Time
	access      map[string]knownAccess
	status      SyncStatus
}

type SyncStatus struct {
	Develop     BranchStatus  `json:"develop"`
	Main        *BranchStatus `json:"main,omitempty"`
	LastAttempt time.Time     `json:"lastAttempt"`
	LastError   string        `json:"lastError,omitempty"`
}

type BranchStatus struct {
	Branch   string    `json:"branch"`
	SHA      string    `json:"sha"`
	SyncedAt time.Time `json:"syncedAt"`
	NotYet   bool      `json:"notYet,omitempty"`
}

type snapshot struct {
	board   Board
	develop taskfiles.Result
	main    *taskfiles.Result
	files   map[string][]byte
	reviews reviews
	docs    map[string][]byte
	assets  map[string][]byte
	links   markdown.Links
	byID    map[string]taskfiles.Task
	heads   map[string]string
}

func NewService(source Source, opts Options) *Service {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	return &Service{source: source, opts: opts, now: time.Now, access: map[string]knownAccess{}}
}

type BoardView struct {
	Board
	Title  string     `json:"title"`
	PRLink string     `json:"prLink,omitempty"`
	Origin Origin     `json:"origin"`
	Sync   SyncStatus `json:"sync"`

	ReadOnly bool `json:"readOnly"`
}

func (s *Service) Board() (BoardView, bool) {
	snap := s.snapshot.Load()
	if snap == nil {
		return BoardView{Title: s.opts.Title, PRLink: s.opts.PRLink, Origin: s.opts.Origin, Sync: s.Status()}, false
	}
	return BoardView{Title: s.opts.Title, PRLink: s.opts.PRLink, Origin: s.opts.Origin, Board: snap.board, Sync: s.Status()}, true
}

func (s *Service) Status() SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := s.status
	if status.Main != nil {
		main := *status.Main
		status.Main = &main
	}
	return status
}

type TaskView struct {
	Card
	Custom []taskfiles.Field `json:"custom"`
	HTML   string            `json:"html"`
}

func (s *Service) Task(id string) (TaskView, error) {
	snap := s.snapshot.Load()
	if snap == nil {
		return TaskView{}, ErrNotFound
	}
	t, ok := snap.byID[id]
	if !ok {
		return TaskView{}, ErrNotFound
	}
	html, err := markdown.Render([]byte(t.Body), t.Path, snap.links)
	if err != nil {
		return TaskView{}, err
	}
	view := TaskView{Custom: orEmpty(t.Custom), HTML: html}
	if card, found := snap.cardByID(id); found {
		view.Card = card
	} else {
		view.Card = archivedCard(t)
	}
	return view, nil
}

func (snap *snapshot) cardByID(id string) (Card, bool) {
	for _, c := range snap.board.Cards {
		if c.ID == id {
			return c, true
		}
	}
	return Card{}, false
}

func (s *Service) Search(query string) []string {
	snap := s.snapshot.Load()
	if snap == nil {
		return []string{}
	}
	return Search(snap.develop.Tasks, query)
}

func (s *Service) Archive() []Card {
	snap := s.snapshot.Load()
	if snap == nil {
		return []Card{}
	}
	out := make([]Card, 0, len(snap.develop.Archived))
	for _, t := range snap.develop.Archived {
		out = append(out, archivedCard(t))
	}
	return out
}

func archivedCard(t taskfiles.Task) Card {
	return Card{
		ID:           t.ID,
		Title:        t.Title,
		Status:       t.Status,
		Owners:       orEmpty(t.Owners),
		Testers:      orEmpty(t.Testers),
		Labels:       labelStrings(t.Labels),
		Prefix:       idPrefix(t.ID),
		Added:        t.Added,
		Claimed:      t.Claimed,
		Done:         t.Done,
		PRs:          orEmpty(t.PRs),
		PRRef:        t.PRRef,
		OpenPRs:      []OpenPR{},
		BlockedBy:    t.BlockedBy,
		BlockedByIDs: []string{},
		Blocks:       []string{},
		References:   []string{},
		ReferencedBy: []string{},
		Path:         t.Path,
	}
}

var markdownLink = regexp.MustCompile(`\]\(([^)\s#]+)`)

func linksTo(t taskfiles.Task, repoPath string) bool {
	for _, m := range markdownLink.FindAllStringSubmatch(t.Body, -1) {
		if path.Clean(path.Join(path.Dir(t.Path), m[1])) == repoPath {
			return true
		}
	}
	return false
}

type DocView struct {
	Path      string   `json:"path"`
	Title     string   `json:"title"`
	HTML      string   `json:"html"`
	Backlinks []string `json:"backlinks"`
}

func (s *Service) Doc(relative string) (DocView, error) {
	snap := s.snapshot.Load()
	if snap == nil {
		return DocView{}, ErrNotFound
	}
	repoPath := path.Clean(path.Join(s.opts.DocsRoot, relative))
	content, ok := snap.docs[repoPath]
	if !ok || !strings.HasPrefix(repoPath, s.opts.DocsRoot+"/") {
		return DocView{}, ErrNotFound
	}
	html, err := markdown.Render(content, repoPath, snap.links)
	if err != nil {
		return DocView{}, err
	}
	name := path.Base(repoPath)
	backlinks := []string{}
	for _, t := range snap.develop.Tasks {
		if linksTo(t, repoPath) {
			backlinks = append(backlinks, t.ID)
		}
	}
	return DocView{
		Path:      strings.TrimPrefix(repoPath, s.opts.DocsRoot+"/"),
		Title:     markdown.Title(content, name),
		HTML:      html,
		Backlinks: backlinks,
	}, nil
}

func (s *Service) Asset(repoPath string) ([]byte, string, error) {
	snap := s.snapshot.Load()
	if snap == nil {
		return nil, "", ErrNotFound
	}
	content, ok := snap.assets[path.Clean(repoPath)]
	if !ok {
		return nil, "", ErrNotFound
	}
	contentType, _ := AssetType(repoPath)
	return content, contentType, nil
}
