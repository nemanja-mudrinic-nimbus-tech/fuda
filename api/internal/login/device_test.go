package login

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type memoryStore struct {
	token   Token
	saved   bool
	deleted bool
}

func (m *memoryStore) Load() (Token, error) {
	if m.token.Access == "" {
		return Token{}, ErrNoStoredToken
	}
	return m.token, nil
}

func (m *memoryStore) Save(token Token) error {
	m.token, m.saved = token, true
	return nil
}

func (m *memoryStore) Delete() error {
	m.token, m.deleted = Token{}, true
	return nil
}

type fakeDeviceGitHub struct {
	server    *httptest.Server
	pending   atomic.Int32
	refreshed atomic.Int32
}

func newFakeDeviceGitHub(t *testing.T) *fakeDeviceGitHub {
	t.Helper()
	f := &fakeDeviceGitHub{}
	f.pending.Store(1)
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/device"):
			_, _ = w.Write([]byte(`{"device_code":"dev","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","interval":5}`))
		case r.Form.Get("grant_type") == "refresh_token":
			f.refreshed.Add(1)
			_, _ = w.Write([]byte(`{"access_token":"refreshed","refresh_token":"next","expires_in":28800}`))
		case f.pending.Add(-1) >= 0:
			_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
		default:
			_, _ = w.Write([]byte(`{"access_token":"user-token","refresh_token":"first","expires_in":28800}`))
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func newDevice(t *testing.T, f *fakeDeviceGitHub, store TokenStore, opened *[]string) *Device {
	t.Helper()
	d := NewDevice(slog.New(slog.DiscardHandler), DeviceConfig{
		ClientID: "id",
		Store:    store,
		Open: func(target string) error {
			*opened = append(*opened, target)
			return nil
		},
	})
	d.app.deviceURL = f.server.URL + "/device"
	d.app.tokenURL = f.server.URL + "/token"
	d.app.client = f.server.Client()
	return d
}

func deviceCall(d *Device, method, target string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	d.Routes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(method, target, nil))
	return w
}

func pollStatus(t *testing.T, d *Device) string {
	t.Helper()
	w := deviceCall(d, "POST", "/auth/github/device/poll")
	var body struct{ Status string }
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.Status
}

func TestDeviceLoginStoresTheTokenInTheStore(t *testing.T) {
	f := newFakeDeviceGitHub(t)
	store := &memoryStore{}
	var opened []string
	d := newDevice(t, f, store, &opened)

	page := deviceCall(d, "GET", "/auth/github/login?return=/github/acme/fuda-tasks/")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "ABCD-1234") {
		t.Fatalf("login page = %d %q, want the user code", page.Code, page.Body.String())
	}
	if len(opened) != 1 || opened[0] != "https://github.com/login/device" {
		t.Fatalf("opened = %v, want the GitHub device page", opened)
	}

	if got := pollStatus(t, d); got != "pending" {
		t.Fatalf("first poll = %q, want pending", got)
	}
	if got := pollStatus(t, d); got != "done" {
		t.Fatalf("second poll = %q, want done", got)
	}
	if !store.saved || store.token.Access != "user-token" {
		t.Fatalf("store = %+v, want the new token saved", store.token)
	}
	token, err := d.Token(nil, httptest.NewRequest("GET", "/", nil))
	if err != nil || token != "user-token" {
		t.Fatalf("Token = %q, %v, want user-token", token, err)
	}
}

func TestDevicePollWithoutLoginFails(t *testing.T) {
	var opened []string
	d := newDevice(t, newFakeDeviceGitHub(t), &memoryStore{}, &opened)
	if got := pollStatus(t, d); got != "failed" {
		t.Fatalf("poll = %q, want failed", got)
	}
}

func TestDeviceTokenNeedsLogin(t *testing.T) {
	var opened []string
	d := newDevice(t, newFakeDeviceGitHub(t), &memoryStore{}, &opened)
	if _, err := d.Token(nil, httptest.NewRequest("GET", "/", nil)); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
}

func TestDeviceTokenRefreshesAnExpiringToken(t *testing.T) {
	f := newFakeDeviceGitHub(t)
	store := &memoryStore{token: Token{Access: "old", Refresh: "r", Expires: time.Now().Add(10 * time.Second)}}
	var opened []string
	d := newDevice(t, f, store, &opened)

	token, err := d.Token(nil, httptest.NewRequest("GET", "/", nil))
	if err != nil || token != "refreshed" {
		t.Fatalf("Token = %q, %v, want refreshed", token, err)
	}
	if store.token.Access != "refreshed" {
		t.Fatalf("store = %+v, want the refreshed token saved", store.token)
	}
}

func TestDeviceTokenForgetsAnExpiredTokenItCannotRefresh(t *testing.T) {
	store := &memoryStore{token: Token{Access: "old", Expires: time.Now().Add(-time.Hour)}}
	var opened []string
	d := newDevice(t, newFakeDeviceGitHub(t), store, &opened)

	if _, err := d.Token(nil, httptest.NewRequest("GET", "/", nil)); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
	if !store.deleted {
		t.Fatal("expired token stayed in the store")
	}
}

func TestDeviceLogoutClearsTheStore(t *testing.T) {
	store := &memoryStore{token: Token{Access: "tok"}}
	var opened []string
	d := newDevice(t, newFakeDeviceGitHub(t), store, &opened)

	if w := deviceCall(d, "POST", "/auth/github/logout"); w.Code != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204", w.Code)
	}
	if _, err := d.Token(nil, httptest.NewRequest("GET", "/", nil)); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn after logout", err)
	}
}

func TestAzureDeviceLoginUsesItsOwnRoutesAndScope(t *testing.T) {
	var scope string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if strings.HasSuffix(r.URL.Path, "/devicecode") {
			scope = r.Form.Get("scope")
			_, _ = w.Write([]byte(`{"device_code":"dev","user_code":"WXYZ","verification_uri":"https://microsoft.com/devicelogin","interval":5}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"azure-token","refresh_token":"r","expires_in":3600}`))
	}))
	t.Cleanup(server.Close)
	store := &memoryStore{}
	d := NewDevice(slog.New(slog.DiscardHandler), DeviceConfig{Host: "azure", Tenant: "organizations", ClientID: "id", Store: store})
	d.app.deviceURL = server.URL + "/devicecode"
	d.app.tokenURL = server.URL + "/token"
	d.app.client = server.Client()

	page := deviceCall(d, "GET", "/auth/azure/login?return=/azure/o/p/fuda-x/")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "WXYZ") || !strings.Contains(page.Body.String(), "/auth/azure/device/poll") {
		t.Fatalf("page: %d %s", page.Code, page.Body)
	}
	if scope != AzureScope {
		t.Errorf("scope %q", scope)
	}
	w := deviceCall(d, "POST", "/auth/azure/device/poll")
	if !strings.Contains(w.Body.String(), "done") || store.token.Access != "azure-token" {
		t.Errorf("poll: %s, stored %q", w.Body, store.token.Access)
	}
}

func TestDeviceLogoutMethodClearsTheStore(t *testing.T) {
	store := &memoryStore{token: Token{Access: "tok"}}
	var opened []string
	d := newDevice(t, newFakeDeviceGitHub(t), store, &opened)

	d.Logout(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil))
	if !store.deleted {
		t.Fatal("token stayed in the store")
	}
}

func newWebDevice(t *testing.T, f *fakeDeviceGitHub) *Device {
	t.Helper()
	d, err := NewWebDevice(slog.New(slog.DiscardHandler), WebDeviceConfig{ClientID: "id", Key: "key"})
	if err != nil {
		t.Fatal(err)
	}
	d.app.deviceURL = f.server.URL + "/device"
	d.app.tokenURL = f.server.URL + "/token"
	d.app.client = f.server.Client()
	return d
}

func webCall(d *Device, method, target, host string, cookies ...*http.Cookie) *http.Response {
	mux := http.NewServeMux()
	d.Routes(mux)
	r := httptest.NewRequest(method, target, nil)
	r.Host = host
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w.Result()
}

func cookieNamed(res *http.Response, name string) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == name && c.MaxAge >= 0 {
			return c
		}
	}
	return nil
}

func TestWebDeviceLoginKeepsTheTokenInASealedCookie(t *testing.T) {
	d := newWebDevice(t, newFakeDeviceGitHub(t))

	page := webCall(d, "GET", "/auth/github/login", "fuda.example.com")
	pending := cookieNamed(page, "fuda_github_device")
	if page.StatusCode != http.StatusOK || pending == nil {
		t.Fatalf("login = %d, pending cookie %v, want the code page and a pending cookie", page.StatusCode, pending)
	}

	first := webCall(d, "POST", "/auth/github/device/poll", "fuda.example.com", pending)
	if cookieNamed(first, "fuda_github") != nil {
		t.Fatal("session cookie set before the person approved")
	}
	done := webCall(d, "POST", "/auth/github/device/poll", "fuda.example.com", pending)
	session := cookieNamed(done, "fuda_github")
	if session == nil {
		t.Fatalf("no session cookie after approval")
	}
	if !session.HttpOnly || !session.Secure || strings.Contains(session.Value, "user-token") {
		t.Errorf("session cookie must be http-only, secure and sealed: %+v", session)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(session)
	if token, err := d.Token(httptest.NewRecorder(), r); err != nil || token != "user-token" {
		t.Fatalf("Token = %q, %v, want user-token", token, err)
	}
	other, err := NewWebDevice(slog.New(slog.DiscardHandler), WebDeviceConfig{ClientID: "id", Key: "key"})
	if err != nil {
		t.Fatal(err)
	}
	if token, err := other.Token(httptest.NewRecorder(), r); err != nil || token != "user-token" {
		t.Fatalf("a restarted server with the same key: Token = %q, %v", token, err)
	}
}

func TestWebDeviceKeepsEachBrowsersLoginApart(t *testing.T) {
	d := newWebDevice(t, newFakeDeviceGitHub(t))
	if webCall(d, "GET", "/auth/github/login", "fuda.example.com").StatusCode != http.StatusOK {
		t.Fatal("login failed")
	}
	res := webCall(d, "POST", "/auth/github/device/poll", "fuda.example.com")
	if !strings.Contains(readBody(res), "failed") {
		t.Fatal("a browser that never started a login got a poll answer")
	}
}

func TestWebDeviceCookiesAreNotSecureOnLocalhost(t *testing.T) {
	d := newWebDevice(t, newFakeDeviceGitHub(t))
	for _, host := range []string{"localhost:5173", "127.0.0.1:8080", "[::1]:8080"} {
		pending := cookieNamed(webCall(d, "GET", "/auth/github/login", host), "fuda_github_device")
		if pending == nil || pending.Secure {
			t.Errorf("%s: cookie = %+v, want a non-secure cookie", host, pending)
		}
	}
	if cookie := cookieNamed(webCall(d, "GET", "/auth/github/login", "192.168.1.5:8080"), "fuda_github_device"); cookie == nil || !cookie.Secure {
		t.Errorf("LAN address: cookie = %+v, want a secure cookie", cookie)
	}
}

func TestWebDeviceLogoutClearsBothCookies(t *testing.T) {
	d := newWebDevice(t, newFakeDeviceGitHub(t))
	res := webCall(d, "POST", "/auth/github/logout", "fuda.example.com")
	cleared := map[string]bool{}
	for _, c := range res.Cookies() {
		cleared[c.Name] = c.MaxAge < 0
	}
	if res.StatusCode != http.StatusNoContent || !cleared["fuda_github"] || !cleared["fuda_github_device"] {
		t.Errorf("logout = %d, cleared %v", res.StatusCode, cleared)
	}
}

func readBody(res *http.Response) string {
	var b strings.Builder
	_, _ = io.Copy(&b, res.Body)
	return b.String()
}
