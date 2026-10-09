package login

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var ErrNoStoredToken = errors.New("no stored token")

type TokenStore interface {
	Load() (Token, error)
	Save(Token) error
	Delete() error
}

type DeviceConfig struct {
	Host     string
	Tenant   string
	ClientID string
	Store    TokenStore
	Open     func(url string) error
}

type pendingDevice struct {
	Code     string    `json:"c"`
	Interval int       `json:"i"`
	Expires  time.Time `json:"x"`
}

type sessions interface {
	token(r *http.Request) (Token, bool)
	save(w http.ResponseWriter, r *http.Request, token Token)
	forget(w http.ResponseWriter, r *http.Request)
	pending(r *http.Request) *pendingDevice
	setPending(w http.ResponseWriter, r *http.Request, pending *pendingDevice)
}

type Device struct {
	app      *oauthApp
	sessions sessions
	open     func(string) error
	log      *slog.Logger
	now      func() time.Time
}

func NewDevice(log *slog.Logger, cfg DeviceConfig) *Device {
	app := newGitHubApp(cfg.ClientID)
	if cfg.Host == "azure" {
		app = newAzureApp(cfg.Tenant, cfg.ClientID, "", "")
	}
	return &Device{
		app:      app,
		sessions: &storeSessions{name: app.name, store: cfg.Store, log: log},
		open:     cfg.Open,
		log:      log,
		now:      time.Now,
	}
}

type WebDeviceConfig struct {
	ClientID string
	Key      string
}

func NewWebDevice(log *slog.Logger, cfg WebDeviceConfig) (*Device, error) {
	sealer, err := newSealer(cfg.Key)
	if err != nil {
		return nil, err
	}
	app := newGitHubApp(cfg.ClientID)
	return &Device{
		app:      app,
		sessions: &cookieSessions{session: "fuda_" + app.name, device: "fuda_" + app.name + "_device", path: "/auth/" + app.name, sealer: sealer},
		log:      log,
		now:      time.Now,
	}, nil
}

func (d *Device) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/"+d.app.name+"/login", d.login)
	mux.HandleFunc("POST /auth/"+d.app.name+"/device/poll", d.poll)
	mux.HandleFunc("POST /auth/"+d.app.name+"/logout", d.logout)
}

func (d *Device) Token(w http.ResponseWriter, r *http.Request) (string, error) {
	token, ok := d.sessions.token(r)
	if !ok {
		return "", ErrNotLoggedIn
	}
	if token.Expires.IsZero() || token.Expires.After(d.now().Add(refreshMargin)) {
		return token.Access, nil
	}
	if token.Refresh == "" {
		d.log.Info(d.app.name + " login expired and cannot be refreshed")
		d.sessions.forget(w, r)
		return "", ErrNotLoggedIn
	}
	next, err := d.app.refresh(r.Context(), token.Refresh)
	if err != nil {
		d.log.Info(d.app.name+" login expired and could not be refreshed", "error", err)
		d.sessions.forget(w, r)
		return "", ErrNotLoggedIn
	}
	d.sessions.save(w, r, next)
	return next.Access, nil
}

type loginPage struct {
	Label    string
	Host     string
	UserCode string
	URL      string
	Interval int
	Return   string
}

var loginTemplate = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Log in to fuda</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>body{font-family:system-ui,sans-serif;max-width:28rem;margin:4rem auto;padding:0 1rem;text-align:center}
code{display:block;font-size:2.25rem;letter-spacing:.2em;margin:1.5rem 0}</style></head>
<body><h1>Log in to fuda</h1>
<p>{{.Label}} opened in your browser. Type this code there:</p>
<code>{{.UserCode}}</code>
<p>Browser did not open? Go to <a href="{{.URL}}">{{.URL}}</a>.</p>
<p id="status">Waiting for you…</p>
<script>
const back = {{.Return}};
let wait = {{.Interval}} * 1000;
async function poll() {
  const res = await fetch('/auth/{{.Host}}/device/poll', { method: 'POST' });
  const body = await res.json();
  if (body.status === 'done') { location.replace(back); return; }
  if (body.status === 'failed') { document.getElementById('status').textContent = body.error; return; }
  if (body.interval) wait = body.interval * 1000;
  setTimeout(poll, wait);
}
setTimeout(poll, wait);
</script></body></html>
`))

func (d *Device) login(w http.ResponseWriter, r *http.Request) {
	code, err := d.app.startDevice(r.Context())
	if err != nil {
		d.log.Warn(d.app.name+" device login failed to start", "error", err)
		http.Error(w, d.app.label+" login failed. Try again.", http.StatusBadGateway)
		return
	}
	interval := max(code.Interval, 5)
	d.sessions.setPending(w, r, &pendingDevice{Code: code.DeviceCode, Interval: interval, Expires: d.now().Add(deviceLifetime)})

	if d.open != nil {
		if err := d.open(code.VerificationURI); err != nil {
			d.log.Warn("open browser for "+d.app.name+" login", "error", err)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = loginTemplate.Execute(w, loginPage{
		Label:    d.app.label,
		Host:     d.app.name,
		UserCode: code.UserCode,
		URL:      code.VerificationURI,
		Interval: interval,
		Return:   safeReturn(r.URL.Query().Get("return")),
	})
}

func (d *Device) poll(w http.ResponseWriter, r *http.Request) {
	pending := d.sessions.pending(r)
	if pending == nil || d.now().After(pending.Expires) {
		d.sessions.setPending(w, r, nil)
		writeStatus(w, map[string]any{"status": "failed", "error": "The login ran out of time. Go back and try again."})
		return
	}
	token, err := d.app.pollDevice(r.Context(), pending.Code)
	switch {
	case errors.Is(err, errAuthorizationPending):
		writeStatus(w, map[string]any{"status": "pending", "interval": pending.Interval})
	case errors.Is(err, errSlowDown):
		pending.Interval += 5
		d.sessions.setPending(w, r, pending)
		writeStatus(w, map[string]any{"status": "pending", "interval": pending.Interval})
	case err != nil:
		d.log.Warn(d.app.name+" device login failed", "error", err)
		d.sessions.setPending(w, r, nil)
		writeStatus(w, map[string]any{"status": "failed", "error": d.app.label + " did not let you in. Go back and try again."})
	default:
		d.sessions.setPending(w, r, nil)
		d.sessions.save(w, r, token)
		writeStatus(w, map[string]any{"status": "done"})
	}
}

func (d *Device) logout(w http.ResponseWriter, r *http.Request) {
	d.Logout(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (d *Device) Logout(w http.ResponseWriter, r *http.Request) {
	d.sessions.forget(w, r)
	d.sessions.setPending(w, r, nil)
}

func writeStatus(w http.ResponseWriter, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func (d *Device) Host() string { return d.app.name }

type storeSessions struct {
	name  string
	store TokenStore
	log   *slog.Logger

	mu      sync.Mutex
	current Token
	waiting *pendingDevice
}

func (s *storeSessions) token(*http.Request) (Token, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current.Access == "" {
		stored, err := s.store.Load()
		if err != nil {
			if !errors.Is(err, ErrNoStoredToken) {
				s.log.Warn("read stored "+s.name+" token", "error", err)
			}
			return Token{}, false
		}
		s.current = stored
	}
	return s.current, true
}

func (s *storeSessions) save(_ http.ResponseWriter, _ *http.Request, token Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = token
	if err := s.store.Save(token); err != nil {
		s.log.Error("store "+s.name+" token", "error", err)
	}
}

func (s *storeSessions) forget(http.ResponseWriter, *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = Token{}
	if err := s.store.Delete(); err != nil {
		s.log.Error("delete stored "+s.name+" token", "error", err)
	}
}

func (s *storeSessions) pending(*http.Request) *pendingDevice {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.waiting
}

func (s *storeSessions) setPending(_ http.ResponseWriter, _ *http.Request, pending *pendingDevice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waiting = pending
}
