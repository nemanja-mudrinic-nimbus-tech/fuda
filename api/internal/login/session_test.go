package login

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type fakeAzure struct {
	server    *httptest.Server
	exchanged atomic.Int32
	refreshed atomic.Int32
	failNext  atomic.Bool
	challenge string
}

func newFakeAzure(t *testing.T) *fakeAzure {
	t.Helper()
	f := &fakeAzure{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		if f.failNext.Load() {
			_, _ = w.Write([]byte(`{"error":"bad_refresh_token","error_description":"expired"}`))
			return
		}
		if r.Form.Get("grant_type") == "refresh_token" {
			n := f.refreshed.Add(1)
			_, _ = w.Write([]byte(`{"access_token":"refreshed-` + strconv.Itoa(int(n)) + `","refresh_token":"next","expires_in":28800}`))
			return
		}
		f.exchanged.Add(1)
		if r.Form.Get("grant_type") != "authorization_code" {
			_, _ = w.Write([]byte(`{"error":"unsupported_grant_type"}`))
			return
		}
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if base64.RawURLEncoding.EncodeToString(sum[:]) != f.challenge {
			_, _ = w.Write([]byte(`{"error":"bad_verification_code"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"user-token","refresh_token":"first","expires_in":28800}`))
	}))
	t.Cleanup(f.server.Close)
	return f
}

func newLogin(t *testing.T, f *fakeAzure) *Web {
	t.Helper()
	g, err := NewAzure(slog.New(slog.DiscardHandler), Config{Tenant: "organizations", ClientID: "id", ClientSecret: "secret", CookieSecret: "cookie-secret", BaseURL: "http://fuda.test"})
	if err != nil {
		t.Fatal(err)
	}
	g.app.tokenURL = f.server.URL
	g.app.client = f.server.Client()
	return g
}

func cookiesOf(res *http.Response) []*http.Cookie { return res.Cookies() }

func call(g *Web, method, target string, cookies ...*http.Cookie) *http.Response {
	mux := http.NewServeMux()
	g.Routes(mux)
	r := httptest.NewRequest(method, target, nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w.Result()
}

func logIn(t *testing.T, g *Web, f *fakeAzure, returnTo string) *http.Cookie {
	t.Helper()
	start := call(g, "GET", "/auth/azure/login?return="+url.QueryEscape(returnTo))
	if start.StatusCode != http.StatusFound {
		t.Fatalf("login: %d", start.StatusCode)
	}
	target, _ := url.Parse(start.Header.Get("Location"))
	query := target.Query()
	if query.Get("code_challenge_method") != "S256" || query.Get("client_id") != "id" || query.Get("redirect_uri") != "http://fuda.test/auth/azure/callback" {
		t.Fatalf("authorize url: %s", target)
	}
	f.challenge = query.Get("code_challenge")

	done := call(g, "GET", "/auth/azure/callback?code=abc&state="+url.QueryEscape(query.Get("state")), cookiesOf(start)...)
	if done.StatusCode != http.StatusFound || done.Header.Get("Location") != returnTo {
		t.Fatalf("callback: %d to %q", done.StatusCode, done.Header.Get("Location"))
	}
	for _, c := range done.Cookies() {
		if c.Name == "fuda_azure" {
			if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || strings.Contains(c.Value, "user-token") {
				t.Errorf("session cookie must be http-only, lax and sealed: %+v", c)
			}
			return c
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func token(g *Web, cookie *http.Cookie) (string, *httptest.ResponseRecorder, error) {
	r := httptest.NewRequest("GET", "/", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	value, err := g.Token(w, r)
	return value, w, err
}

func TestLoginKeepsTheTokenInAnEncryptedCookie(t *testing.T) {
	f := newFakeAzure(t)
	g := newLogin(t, f)
	cookie := logIn(t, g, f, "/azure/o/p/fuda-x/board")

	got, _, err := token(g, cookie)
	if err != nil || got != "user-token" {
		t.Errorf("got %q, %v", got, err)
	}
	if _, _, err := token(g, nil); !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("no cookie: %v", err)
	}
	if _, _, err := token(g, &http.Cookie{Name: "fuda_azure", Value: "garbage"}); !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("garbage cookie: %v", err)
	}
}

func TestCallbackRefusesAForeignState(t *testing.T) {
	f := newFakeAzure(t)
	g := newLogin(t, f)
	start := call(g, "GET", "/auth/azure/login")
	res := call(g, "GET", "/auth/azure/callback?code=abc&state=other", cookiesOf(start)...)
	if res.StatusCode != http.StatusBadRequest || f.exchanged.Load() != 0 {
		t.Errorf("status %d, exchanges %d", res.StatusCode, f.exchanged.Load())
	}
	if res := call(g, "GET", "/auth/azure/callback?code=abc&state=x"); res.StatusCode != http.StatusBadRequest {
		t.Errorf("no flow cookie: %d", res.StatusCode)
	}
}

func TestLoginReturnsOnlyToThisSite(t *testing.T) {
	f := newFakeAzure(t)
	g := newLogin(t, f)
	for _, bad := range []string{"https://evil.test/", "//evil.test", "/\\evil.test", "/\t/evil.test"} {
		start := call(g, "GET", "/auth/azure/login?return="+url.QueryEscape(bad))
		query, _ := url.Parse(start.Header.Get("Location"))
		f.challenge = query.Query().Get("code_challenge")
		done := call(g, "GET", "/auth/azure/callback?code=abc&state="+url.QueryEscape(query.Query().Get("state")), cookiesOf(start)...)
		if done.Header.Get("Location") != "/" {
			t.Errorf("%s redirected to %s", bad, done.Header.Get("Location"))
		}
	}
}

func TestAnExpiredTokenIsRefreshedOnce(t *testing.T) {
	f := newFakeAzure(t)
	g := newLogin(t, f)
	cookie := logIn(t, g, f, "/")
	g.now = func() time.Time { return time.Now().Add(9 * time.Hour) }

	first, w, err := token(g, cookie)
	if err != nil || first != "refreshed-1" {
		t.Fatalf("got %q, %v", first, err)
	}
	if len(w.Result().Cookies()) == 0 {
		t.Error("the refreshed token was not stored")
	}
	second, _, err := token(g, cookie)
	if err != nil || second != "refreshed-1" || f.refreshed.Load() != 1 {
		t.Errorf("a request still holding the old cookie must reuse the refresh: %q, %v, %d refreshes", second, err, f.refreshed.Load())
	}
}

func TestAFailedRefreshSendsTheUserBackToLogin(t *testing.T) {
	f := newFakeAzure(t)
	g := newLogin(t, f)
	cookie := logIn(t, g, f, "/")
	g.now = func() time.Time { return time.Now().Add(9 * time.Hour) }
	f.failNext.Store(true)

	_, w, err := token(g, cookie)
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("got %v", err)
	}
	cleared := false
	for _, c := range w.Result().Cookies() {
		cleared = cleared || (c.Name == "fuda_azure" && c.MaxAge < 0)
	}
	if !cleared {
		t.Error("the dead login cookie was kept")
	}
}

func TestLogoutClearsTheCookie(t *testing.T) {
	g := newLogin(t, newFakeAzure(t))
	res := call(g, "POST", "/auth/azure/logout")
	if res.StatusCode != http.StatusNoContent || len(res.Cookies()) == 0 || res.Cookies()[0].MaxAge >= 0 {
		t.Errorf("%d %+v", res.StatusCode, res.Cookies())
	}
}

func TestLogoutMethodClearsOnlyItsOwnHostCookie(t *testing.T) {
	g := newLogin(t, newFakeAzure(t))
	w := httptest.NewRecorder()
	g.Logout(w, httptest.NewRequest("POST", "/", nil))
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "fuda_azure" || cookies[0].MaxAge >= 0 {
		t.Errorf("%+v", cookies)
	}
}
