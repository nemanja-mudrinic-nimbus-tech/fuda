package board

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/singleflight"
)

type BoardID struct {
	Host string
	Repo string
}

var segment = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._ -]*$`)

func (id BoardID) Path() string {
	return (&url.URL{Path: "/" + id.Host + "/" + id.Repo}).EscapedPath()
}

func (id BoardID) valid() bool {
	if id.Host == "" || id.Repo == "" {
		return false
	}
	for part := range strings.SplitSeq(id.Repo, "/") {
		if !segment.MatchString(part) || part == ".." {
			return false
		}
	}
	return true
}

type Boards struct {
	open func(BoardID) (*Service, error)
	list func(ctx context.Context, host string) ([]BoardID, error)
	log  *slog.Logger

	opening  singleflight.Group
	mu       sync.Mutex
	services map[BoardID]*Service
}

func NewBoards(log *slog.Logger, open func(BoardID) (*Service, error), list func(ctx context.Context, host string) ([]BoardID, error)) *Boards {
	return &Boards{open: open, list: list, log: log, services: map[BoardID]*Service{}}
}

func (b *Boards) List(ctx context.Context, host string) ([]BoardID, error) {
	return b.list(ctx, host)
}

func (b *Boards) Get(ctx context.Context, id BoardID) (*Service, error) {
	if !id.valid() {
		return nil, ErrNotFound
	}
	value, err, _ := b.opening.Do(id.Path(), func() (any, error) {
		if service, ok := b.cached(id); ok {
			return service, nil
		}
		return b.start(id)
	})
	if err != nil {
		return nil, err
	}
	service := value.(*Service)
	if err := service.Poll(ctx); err != nil {
		if _, ready := service.Board(); !ready {
			b.forget(id)
		}
		return nil, err
	}
	return service, nil
}

func (b *Boards) cached(id BoardID) (*Service, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	service, ok := b.services[id]
	return service, ok
}

func (b *Boards) start(id BoardID) (*Service, error) {
	service, err := b.open(id)
	if err != nil {
		return nil, err
	}
	if err := service.Restore(); err != nil {
		b.log.Warn("the disk cache could not be read; starting empty", "board", id.Path(), "error", err)
	}
	b.mu.Lock()
	b.services[id] = service
	b.mu.Unlock()
	return service, nil
}

func (b *Boards) forget(id BoardID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.services, id)
}
