package login

import (
	"context"
	"errors"
	"sync"
	"time"
)

type refreshed struct {
	token Token
	at    time.Time
}

type refreshMemo struct {
	app *provider
	now func() time.Time

	mu     sync.Mutex
	recent map[string]refreshed
}

func newRefreshMemo(app *provider, now func() time.Time) *refreshMemo {
	return &refreshMemo{app: app, now: now, recent: map[string]refreshed{}}
}

func (m *refreshMemo) refresh(ctx context.Context, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("no refresh token")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, entry := range m.recent {
		if m.now().Sub(entry.at) > memoLifetime {
			delete(m.recent, key)
		}
	}
	if entry, ok := m.recent[refreshToken]; ok {
		return entry.token, nil
	}
	token, err := m.app.refresh(ctx, refreshToken)
	if err != nil {
		return Token{}, err
	}
	m.recent[refreshToken] = refreshed{token: token, at: m.now()}
	return token, nil
}
