package config

import (
	"strings"
	"testing"

	"fuda/internal/board"
)

var secret = strings.Repeat("c", minCookieSecretLength)

func TestHostsTakeAList(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "github,azure")
	t.Setenv("FUDA_GITHUB_CLIENT_ID", "g")
	t.Setenv("FUDA_GITHUB_APP_SLUG", "fuda")
	t.Setenv("FUDA_AZURE_CLIENT_ID", "a")
	t.Setenv("FUDA_AZURE_CLIENT_SECRET", "a")
	t.Setenv("FUDA_COOKIE_SECRET", secret)
	t.Setenv("FUDA_BASE_URL", "http://fuda.test")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Has(board.HostGitHub) || !cfg.Has(board.HostAzure) || cfg.Has(board.HostLocal) {
		t.Errorf("hosts: %v", cfg.Hosts)
	}
}

func TestEachHostNeedsItsOwnSettings(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "github,azure")
	t.Setenv("FUDA_GITHUB_CLIENT_ID", "g")
	t.Setenv("FUDA_GITHUB_APP_SLUG", "fuda")
	t.Setenv("FUDA_COOKIE_SECRET", secret)
	t.Setenv("FUDA_BASE_URL", "http://fuda.test")
	if _, err := Load(); err == nil {
		t.Error("azure without its client settings passed")
	}
}

func TestLocalStandsAlone(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "local,github")
	t.Setenv("FUDA_LOCAL_PATH", "/tmp/x")
	if _, err := Load(); err == nil {
		t.Error("local mixed with github passed")
	}
}

func TestGitHubNeedsTheClientIdAndTheAppSlug(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "github")
	t.Setenv("FUDA_GITHUB_CLIENT_ID", "g")
	if _, err := Load(); err == nil {
		t.Error("github without the app slug passed")
	}
	t.Setenv("FUDA_GITHUB_APP_SLUG", "fuda")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestRemovedVarsAreReported(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "github")
	t.Setenv("FUDA_GITHUB_CLIENT_ID", "g")
	t.Setenv("FUDA_GITHUB_APP_SLUG", "fuda")
	t.Setenv("FUDA_GITHUB_CLIENT_SECRET", "s")
	t.Setenv("FUDA_COOKIE_SECRET", secret)
	t.Setenv("FUDA_BASE_URL", "http://fuda.test")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.IgnoredVars(); len(got) != 3 {
		t.Errorf("ignored = %v, want the three unused vars", got)
	}
}

func TestAzureKeepsTheCookieVars(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "github,azure")
	t.Setenv("FUDA_GITHUB_CLIENT_ID", "g")
	t.Setenv("FUDA_GITHUB_APP_SLUG", "fuda")
	t.Setenv("FUDA_AZURE_CLIENT_ID", "a")
	t.Setenv("FUDA_AZURE_CLIENT_SECRET", "a")
	t.Setenv("FUDA_COOKIE_SECRET", secret)
	t.Setenv("FUDA_BASE_URL", "http://fuda.test")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.IgnoredVars(); len(got) != 0 {
		t.Errorf("ignored = %v, want none", got)
	}
}

func TestTheOldSourceVarStillSetsTheHosts(t *testing.T) {
	t.Setenv("FUDA_SOURCE", "local")
	t.Setenv("FUDA_LOCAL_PATH", "/tmp/x")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Has(board.HostLocal) || len(cfg.Hosts) != 1 {
		t.Errorf("hosts: %v", cfg.Hosts)
	}
}

func TestAShortCookieSecretIsRefused(t *testing.T) {
	t.Setenv("FUDA_HOSTS", "azure")
	t.Setenv("FUDA_AZURE_CLIENT_ID", "a")
	t.Setenv("FUDA_AZURE_CLIENT_SECRET", "a")
	t.Setenv("FUDA_COOKIE_SECRET", "short")
	t.Setenv("FUDA_BASE_URL", "http://fuda.test")
	if _, err := Load(); err == nil {
		t.Error("a short cookie secret passed")
	}
}

func TestDesktopFallsBackToTheBuildValues(t *testing.T) {
	cfg, err := LoadDesktop(DesktopDefaults{GitHubClientID: "g", GitHubAppSlug: "fuda", CacheDir: "/cache/fuda"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GitHubClientID != "g" || cfg.CacheDir != "/cache/fuda" || !cfg.LocalWrites {
		t.Errorf("config: %+v", cfg)
	}
	if !cfg.Has(board.HostGitHub) || cfg.Has(board.HostAzure) {
		t.Errorf("hosts: %v", cfg.Hosts)
	}
}

func TestDesktopEnvironmentWinsOverTheBuildValues(t *testing.T) {
	t.Setenv("FUDA_GITHUB_CLIENT_ID", "from-env")
	t.Setenv("FUDA_AZURE_CLIENT_ID", "az")
	cfg, err := LoadDesktop(DesktopDefaults{GitHubClientID: "g", GitHubAppSlug: "fuda", CacheDir: "/cache/fuda"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GitHubClientID != "from-env" || !cfg.Has(board.HostAzure) {
		t.Errorf("config: %+v", cfg)
	}
}

func TestDesktopNeedsTheGitHubBuildValues(t *testing.T) {
	if _, err := LoadDesktop(DesktopDefaults{}); err == nil {
		t.Error("a desktop without the GitHub values passed")
	}
}
