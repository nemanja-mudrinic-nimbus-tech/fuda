package login

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeyIsCreatedPrivateAndReused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache", "cookie.key")
	first, err := LoadOrCreateKey(path)
	if err != nil || first == "" {
		t.Fatalf("first = %q, %v", first, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("key file mode = %v, %v, want 0600", info.Mode().Perm(), err)
	}
	second, err := LoadOrCreateKey(path)
	if err != nil || second != first {
		t.Fatalf("second = %q, %v, want the saved key %q", second, err, first)
	}
}
