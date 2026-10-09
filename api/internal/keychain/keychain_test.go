package keychain

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"

	"fuda/internal/login"
)

func TestStoreRoundTripsATokenThroughTheKeychain(t *testing.T) {
	keyring.MockInit()
	store := New("github")

	if _, err := store.Load(); !errors.Is(err, login.ErrNoStoredToken) {
		t.Fatalf("empty Load err = %v, want ErrNoStoredToken", err)
	}
	if err := store.Save(login.Token{Access: "a", Refresh: "r"}); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got.Access != "a" || got.Refresh != "r" {
		t.Fatalf("Load = %+v, %v, want the saved token", got, err)
	}
	if err := store.Delete(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); !errors.Is(err, login.ErrNoStoredToken) {
		t.Fatalf("Load after Delete err = %v, want ErrNoStoredToken", err)
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("second Delete = %v, want nil", err)
	}
}
