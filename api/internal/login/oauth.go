package login

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	errAuthorizationPending = errors.New("authorization pending")
	errSlowDown             = errors.New("the login server asked to poll slower")
)

type Token struct {
	Access  string    `json:"a"`
	Refresh string    `json:"r,omitempty"`
	Expires time.Time `json:"e,omitzero"`
}

type oauthApp struct {
	name         string
	label        string
	scope        string
	clientID     string
	clientSecret string
	redirectURL  string
	authorizeURL string
	tokenURL     string
	deviceURL    string
	client       *http.Client
	now          func() time.Time
}

func newGitHubApp(clientID string) *oauthApp {
	return &oauthApp{
		name:      "github",
		label:     "GitHub",
		clientID:  clientID,
		tokenURL:  "https://github.com/login/oauth/access_token",
		deviceURL: "https://github.com/login/device/code",
		client:    &http.Client{Timeout: 30 * time.Second},
		now:       time.Now,
	}
}

func newAzureApp(tenant, clientID, clientSecret, redirectURL string) *oauthApp {
	authority := "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0"
	return &oauthApp{
		name:         "azure",
		label:        "Azure DevOps",
		scope:        AzureScope,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		authorizeURL: authority + "/authorize",
		tokenURL:     authority + "/token",
		deviceURL:    authority + "/devicecode",
		client:       &http.Client{Timeout: 30 * time.Second},
		now:          time.Now,
	}
}

func (g *oauthApp) authorize(state, challenge string) string {
	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {g.clientID},
		"redirect_uri":          {g.redirectURL},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	if g.scope != "" {
		query.Set("scope", g.scope)
	}
	return g.authorizeURL + "?" + query.Encode()
}

func (g *oauthApp) exchange(ctx context.Context, code, verifier string) (Token, error) {
	return g.token(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {g.redirectURL},
		"code_verifier": {verifier},
	})
}

func (g *oauthApp) refresh(ctx context.Context, refreshToken string) (Token, error) {
	return g.token(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	})
}

func (g *oauthApp) token(ctx context.Context, form url.Values) (Token, error) {
	form.Set("client_id", g.clientID)
	if g.scope != "" {
		form.Set("scope", g.scope)
	}
	if g.clientSecret != "" {
		form.Set("client_secret", g.clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := g.client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer func() { _ = res.Body.Close() }()
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		Description  string `json:"error_description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return Token{}, fmt.Errorf("%s token response: %w", g.label, err)
	}
	switch body.Error {
	case "authorization_pending":
		return Token{}, errAuthorizationPending
	case "slow_down":
		return Token{}, errSlowDown
	}
	if body.Error != "" {
		return Token{}, fmt.Errorf("%s refused the token request: %s: %s", g.label, body.Error, body.Description)
	}
	if body.AccessToken == "" {
		return Token{}, fmt.Errorf("%s sent no access token", g.label)
	}
	token := Token{Access: body.AccessToken, Refresh: body.RefreshToken}
	if body.ExpiresIn > 0 {
		token.Expires = g.now().Add(time.Duration(body.ExpiresIn) * time.Second)
	}
	return token, nil
}

type deviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
	Error           string `json:"error"`
	Description     string `json:"error_description"`
}

func (g *oauthApp) startDevice(ctx context.Context) (deviceCode, error) {
	form := url.Values{"client_id": {g.clientID}}
	if g.scope != "" {
		form.Set("scope", g.scope)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.deviceURL, strings.NewReader(form.Encode()))
	if err != nil {
		return deviceCode{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := g.client.Do(req)
	if err != nil {
		return deviceCode{}, err
	}
	defer func() { _ = res.Body.Close() }()
	var code deviceCode
	if err := json.NewDecoder(res.Body).Decode(&code); err != nil {
		return deviceCode{}, fmt.Errorf("%s device response: %w", g.label, err)
	}
	if code.Error != "" {
		return deviceCode{}, fmt.Errorf("%s refused the device login: %s: %s", g.label, code.Error, code.Description)
	}
	if code.DeviceCode == "" || code.UserCode == "" {
		return deviceCode{}, fmt.Errorf("%s sent no device code", g.label)
	}
	return code, nil
}

func (g *oauthApp) pollDevice(ctx context.Context, code string) (Token, error) {
	return g.token(ctx, url.Values{
		"device_code": {code},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	})
}

const AzureScope = "499b84ac-1321-427f-aa17-267ca6975798/user_impersonation offline_access"
