package login

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ErrNotLoggedIn = errors.New("not logged in")

const (
	flowLifetime  = 10 * time.Minute
	refreshMargin = time.Minute
	memoLifetime  = time.Minute
)

type Config struct {
	Tenant       string
	ClientID     string
	ClientSecret string
	CookieSecret string
	BaseURL      string
}

type Web struct {
	session string
	flow    string
	app     *oauthApp
	sealer  sealer
	secure  bool
	log     *slog.Logger
	now     func() time.Time

	mu     sync.Mutex
	recent map[string]refreshed
}

type refreshed struct {
	token Token
	at    time.Time
}

type flow struct {
	State    string `json:"s"`
	Verifier string `json:"v"`
	Return   string `json:"r"`
}

func NewAzure(log *slog.Logger, cfg Config) (*Web, error) {
	sealer, err := newSealer(cfg.CookieSecret)
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(cfg.BaseURL, "/")
	app := newAzureApp(cfg.Tenant, cfg.ClientID, cfg.ClientSecret, "")
	app.redirectURL = base + "/auth/" + app.name + "/callback"
	return &Web{
		session: "fuda_" + app.name,
		flow:    "fuda_" + app.name + "_flow",
		app:     app,
		sealer:  sealer,
		secure:  strings.HasPrefix(base, "https://"),
		log:     log,
		now:     time.Now,
		recent:  map[string]refreshed{},
	}, nil
}

func (g *Web) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/"+g.app.name+"/login", g.login)
	mux.HandleFunc("GET /auth/"+g.app.name+"/callback", g.callback)
	mux.HandleFunc("POST /auth/"+g.app.name+"/logout", g.logout)
}

func (g *Web) Token(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(g.session)
	if err != nil {
		return "", ErrNotLoggedIn
	}
	var token Token
	if err := g.sealer.open(cookie.Value, &token); err != nil || token.Access == "" {
		g.clear(w, g.session, "/")
		return "", ErrNotLoggedIn
	}
	if token.Expires.IsZero() || token.Expires.After(g.now().Add(refreshMargin)) {
		return token.Access, nil
	}
	next, err := g.refreshOnce(r, token.Refresh)
	if err != nil {
		g.log.Info("login expired and could not be refreshed", "error", err)
		g.clear(w, g.session, "/")
		return "", ErrNotLoggedIn
	}
	g.storeSession(w, next)
	return next.Access, nil
}

func (g *Web) refreshOnce(r *http.Request, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("no refresh token")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for key, entry := range g.recent {
		if g.now().Sub(entry.at) > memoLifetime {
			delete(g.recent, key)
		}
	}
	if entry, ok := g.recent[refreshToken]; ok {
		return entry.token, nil
	}
	token, err := g.app.refresh(r.Context(), refreshToken)
	if err != nil {
		return Token{}, err
	}
	g.recent[refreshToken] = refreshed{token: token, at: g.now()}
	return token, nil
}

func (g *Web) login(w http.ResponseWriter, r *http.Request) {
	state, verifier := randomString(), randomString()
	value, err := g.sealer.seal(flow{State: state, Verifier: verifier, Return: safeReturn(r.URL.Query().Get("return"))})
	if err != nil {
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}
	g.set(w, g.flow, "/auth/"+g.app.name, value, flowLifetime)
	challenge := sha256.Sum256([]byte(verifier))
	http.Redirect(w, r, g.app.authorize(state, base64.RawURLEncoding.EncodeToString(challenge[:])), http.StatusFound)
}

func (g *Web) callback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(g.flow)
	var started flow
	if err != nil || g.sealer.open(cookie.Value, &started) != nil || started.State == "" || started.State != r.URL.Query().Get("state") {
		http.Error(w, "The login did not start here. Open fuda and try again.", http.StatusBadRequest)
		return
	}
	g.clear(w, g.flow, "/auth/"+g.app.name)
	if r.URL.Query().Get("error") != "" {
		http.Error(w, g.app.label+" did not let you in: "+r.URL.Query().Get("error"), http.StatusForbidden)
		return
	}
	token, err := g.app.exchange(r.Context(), r.URL.Query().Get("code"), started.Verifier)
	if err != nil {
		g.log.Warn(g.app.label+" login failed", "error", err)
		http.Error(w, g.app.label+" login failed. Try again.", http.StatusBadGateway)
		return
	}
	g.storeSession(w, token)
	http.Redirect(w, r, started.Return, http.StatusFound)
}

func (g *Web) logout(w http.ResponseWriter, r *http.Request) {
	g.Logout(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (g *Web) Logout(w http.ResponseWriter, _ *http.Request) {
	g.clear(w, g.session, "/")
}

func (g *Web) storeSession(w http.ResponseWriter, token Token) {
	value, err := g.sealer.seal(token)
	if err != nil {
		g.log.Error("seal login cookie", "error", err)
		return
	}
	lifetime := 180 * 24 * time.Hour
	g.set(w, g.session, "/", value, lifetime)
}

func (g *Web) set(w http.ResponseWriter, name, path, value string, lifetime time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: path, MaxAge: int(lifetime.Seconds()),
		HttpOnly: true, Secure: g.secure, SameSite: http.SameSiteLaxMode,
	})
}

func (g *Web) clear(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Path: path, MaxAge: -1, HttpOnly: true, Secure: g.secure, SameSite: http.SameSiteLaxMode,
	})
}

func safeReturn(target string) string {
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") || strings.ContainsAny(target, "\\\t\r\n") {
		return "/"
	}
	return target
}

func randomString() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func (g *Web) Host() string { return g.app.name }
