package board

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type Host string

const (
	HostGitHub Host = "github"
	HostAzure  Host = "azure"
	HostLocal  Host = "local"
)

type ID struct {
	Host Host
	Repo string
}

var segment = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._ -]*$`)

func (id ID) Path() string {
	return (&url.URL{Path: "/" + string(id.Host) + "/" + id.Repo}).EscapedPath()
}

func (id ID) valid() bool {
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

type Registry struct {
	open func(ID) (*Service, error)
	list func(ctx context.Context, host Host) ([]ID, error)
	log  *slog.Logger

	installs map[Host]AppInstall

	now func() time.Time

	opening  singleflight.Group
	mu       sync.Mutex
	services map[ID]*Service
	gate     map[gateKey]time.Time
}

const gateTTL = 3 * time.Second

type gateKey struct {
	id    ID
	token string
}

type AppInstall struct {
	URL       string
	Installed func(ctx context.Context) (bool, error)
}

func NewRegistry(log *slog.Logger, open func(ID) (*Service, error), list func(ctx context.Context, host Host) ([]ID, error), installs map[Host]AppInstall) *Registry {
	return &Registry{open: open, list: list, log: log, installs: installs, now: time.Now, gate: map[gateKey]time.Time{}, services: map[ID]*Service{}}
}

func (b *Registry) InstallURL(host Host) string {
	return b.installs[host].URL
}

func (b *Registry) AppInstalled(ctx context.Context, host Host) *bool {
	install, ok := b.installs[host]
	if !ok || install.Installed == nil {
		return nil
	}
	installed, err := install.Installed(ctx)
	if err != nil {
		b.log.Warn("could not check the app installation", "host", host, "error", err)
		return nil
	}
	return &installed
}

func (b *Registry) List(ctx context.Context, host Host) ([]ID, error) {
	return b.list(ctx, host)
}

func (b *Registry) Get(ctx context.Context, id ID) (*Service, error) {
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
	key := gateKey{id: id, token: tokenDigest(ctx)}
	if b.gateOpen(key) {
		return service, nil
	}
	if err := service.Poll(ctx); err != nil {
		if _, ready := service.Board(); !ready {
			b.forget(id)
		}
		return nil, err
	}
	b.openGate(key)
	return service, nil
}

func (b *Registry) gateOpen(key gateKey) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	checked, ok := b.gate[key]
	return ok && b.now().Sub(checked) < gateTTL
}

func (b *Registry) openGate(key gateKey) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for old, checked := range b.gate {
		if b.now().Sub(checked) >= gateTTL {
			delete(b.gate, old)
		}
	}
	b.gate[key] = b.now()
}

func (b *Registry) cached(id ID) (*Service, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	service, ok := b.services[id]
	return service, ok
}

func (b *Registry) start(id ID) (*Service, error) {
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

func (b *Registry) forget(id ID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.services, id)
}
