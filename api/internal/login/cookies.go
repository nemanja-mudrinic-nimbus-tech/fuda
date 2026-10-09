package login

import (
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	sessionLifetime = 180 * 24 * time.Hour
	deviceLifetime  = 15 * time.Minute
)

type cookieSessions struct {
	session string
	device  string
	path    string
	sealer  sealer
}

func (c *cookieSessions) token(r *http.Request) (Token, bool) {
	cookie, err := r.Cookie(c.session)
	if err != nil {
		return Token{}, false
	}
	var token Token
	if err := c.sealer.open(cookie.Value, &token); err != nil || token.Access == "" {
		return Token{}, false
	}
	return token, true
}

func (c *cookieSessions) save(w http.ResponseWriter, r *http.Request, token Token) {
	value, err := c.sealer.seal(token)
	if err != nil {
		return
	}
	setCookie(w, r, c.session, "/", value, sessionLifetime)
}

func (c *cookieSessions) forget(w http.ResponseWriter, r *http.Request) {
	setCookie(w, r, c.session, "/", "", -1)
}

func (c *cookieSessions) pending(r *http.Request) *pendingDevice {
	cookie, err := r.Cookie(c.device)
	if err != nil {
		return nil
	}
	var pending pendingDevice
	if err := c.sealer.open(cookie.Value, &pending); err != nil || pending.Code == "" {
		return nil
	}
	return &pending
}

func (c *cookieSessions) setPending(w http.ResponseWriter, r *http.Request, pending *pendingDevice) {
	if pending == nil {
		setCookie(w, r, c.device, c.path, "", -1)
		return
	}
	value, err := c.sealer.seal(pending)
	if err != nil {
		return
	}
	setCookie(w, r, c.device, c.path, value, deviceLifetime)
}

func setCookie(w http.ResponseWriter, r *http.Request, name, path, value string, lifetime time.Duration) {
	maxAge := int(lifetime.Seconds())
	if lifetime < 0 {
		maxAge = -1
	}
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: path, MaxAge: maxAge,
		HttpOnly: true, Secure: !isLocalhost(r.Host), SameSite: http.SameSiteLaxMode,
	})
}

func isLocalhost(hostport string) bool {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = strings.Trim(hostport, "[]")
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
