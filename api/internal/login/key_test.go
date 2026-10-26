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

func TestAnEmptyKeyFileIsFilled(t *testing.T) {
	for _, content := range []string{"", " \n"} {
		path := filepath.Join(t.TempDir(), "cookie.key")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		key, err := LoadOrCreateKey(path)
		if err != nil || key == "" {
			t.Fatalf("content %q: key = %q, %v", content, key, err)
		}
		again, err := LoadOrCreateKey(path)
		if err != nil || again != key {
			t.Fatalf("content %q: again = %q, %v, want %q", content, again, err, key)
		}
	}
}
