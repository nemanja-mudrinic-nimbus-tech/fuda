package version

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func readRepoFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func versionFile(t *testing.T) string {
	t.Helper()

	return strings.TrimSpace(readRepoFile(t, "VERSION"))
}

func TestBoardFormatIsTheMajorMinorOfTheVersionFile(t *testing.T) {
	version := versionFile(t)
	if !regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`).MatchString(version) {
		t.Fatalf("VERSION %q is not X.Y.Z or X.Y.Z-suffix", version)
	}

	parts := strings.SplitN(version, ".", 3)
	if got := parts[0] + "." + parts[1]; got != BoardFormat {
		t.Errorf("VERSION %s has Board format %s, but BoardFormat is %s", version, got, BoardFormat)
	}
}

func TestClientPackageCarriesTheVersionFile(t *testing.T) {
	var pkg struct{ Version string }
	if err := json.Unmarshal([]byte(readRepoFile(t, "client/package.json")), &pkg); err != nil {
		t.Fatal(err)
	}

	if want := versionFile(t); pkg.Version != want {
		t.Errorf("client/package.json version is %q, VERSION is %q", pkg.Version, want)
	}
}

func TestGuidePagesNameTheBoardFormatOfTheConstant(t *testing.T) {
	pages, err := filepath.Glob(filepath.Join("..", "guide", "pages", "*.md"))
	if err != nil || len(pages) == 0 {
		t.Fatalf("guide pages: %v %v", pages, err)
	}

	mention := regexp.MustCompile(`Board format (\d+\.\d+)`)
	for _, page := range pages {
		content, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}

		for _, match := range mention.FindAllStringSubmatch(string(content), -1) {
			if match[1] != BoardFormat {
				t.Errorf("%s names Board format %s, BoardFormat is %s", filepath.Base(page), match[1], BoardFormat)
			}
		}
	}
}
