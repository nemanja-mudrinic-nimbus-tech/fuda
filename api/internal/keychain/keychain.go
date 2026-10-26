package keychain

import (
	"encoding/json"
	"errors"

	"github.com/zalando/go-keyring"

	"fuda/internal/login"
)

const service = "fuda"

type Store struct {
	account string
}

func New(account string) Store {
	return Store{account: account}
}

func (s Store) Load() (login.Token, error) {
	raw, err := keyring.Get(service, s.account)
	if errors.Is(err, keyring.ErrNotFound) {
		return login.Token{}, login.ErrNoStoredToken
	}
	if err != nil {
		return login.Token{}, err
	}
	var token login.Token
	if err := json.Unmarshal([]byte(raw), &token); err != nil {
		return login.Token{}, login.ErrNoStoredToken
	}
	return token, nil
}

func (s Store) Save(token login.Token) error {
	raw, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keyring.Set(service, s.account, string(raw))
}

func (s Store) Delete() error {
	err := keyring.Delete(service, s.account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
