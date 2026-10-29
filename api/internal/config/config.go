package config

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"

	"fuda/internal/board"
)

const minCookieSecretLength = 32

type Config struct {
	Addr  string `env:"FUDA_ADDR" envDefault:":8080"`
	Title string `env:"FUDA_TITLE"`

	BaseURL      string `env:"FUDA_BASE_URL"`
	CookieSecret string `env:"FUDA_COOKIE_SECRET"`

	Hosts       []board.Host `env:"FUDA_HOSTS"`
	LocalPath   string       `env:"FUDA_LOCAL_PATH"`
	LocalWrites bool

	GitHubClientID string `env:"FUDA_GITHUB_CLIENT_ID"`
	GitHubAppSlug  string `env:"FUDA_GITHUB_APP_SLUG"`

	AzureTenant       string `env:"FUDA_AZURE_TENANT" envDefault:"organizations"`
	AzureClientID     string `env:"FUDA_AZURE_CLIENT_ID"`
	AzureClientSecret string `env:"FUDA_AZURE_CLIENT_SECRET"`

	GitHubBoards []string `env:"FUDA_GITHUB_BOARDS"`
	AzureBoards  []string `env:"FUDA_AZURE_BOARDS"`

	WatchMain    bool          `env:"FUDA_WATCH_MAIN" envDefault:"false"`
	SyncCooldown time.Duration `env:"FUDA_SYNC_COOLDOWN" envDefault:"30s"`
	CacheDir     string        `env:"FUDA_CACHE_DIR" envDefault:"./.fuda-cache"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	if len(cfg.Hosts) == 0 {
		cfg.Hosts = hostsFromOldVar()
	}
	cfg.GitHubBoards = boardNames(cfg.GitHubBoards)
	cfg.AzureBoards = boardNames(cfg.AzureBoards)
	return cfg, cfg.validate()
}

type DesktopDefaults struct {
	GitHubClientID string
	GitHubAppSlug  string
	AzureClientID  string
	GitHubBoards   string
	AzureBoards    string
	CacheDir       string
}

func LoadDesktop(defaults DesktopDefaults) (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	cfg.GitHubClientID = cmp.Or(cfg.GitHubClientID, defaults.GitHubClientID)
	cfg.GitHubAppSlug = cmp.Or(cfg.GitHubAppSlug, defaults.GitHubAppSlug)
	cfg.AzureClientID = cmp.Or(cfg.AzureClientID, defaults.AzureClientID)
	cfg.GitHubBoards = boardNames(strings.Split(cmp.Or(strings.Join(cfg.GitHubBoards, ","), defaults.GitHubBoards), ","))
	cfg.AzureBoards = boardNames(strings.Split(cmp.Or(strings.Join(cfg.AzureBoards, ","), defaults.AzureBoards), ","))
	if err := cfg.validateBoards(); err != nil {
		return Config{}, err
	}
	if _, set := os.LookupEnv("FUDA_CACHE_DIR"); !set {
		cfg.CacheDir = defaults.CacheDir
	}
	if cfg.GitHubClientID == "" || cfg.GitHubAppSlug == "" {
		return Config{}, fmt.Errorf("this build has no GitHub client id or app slug")
	}
	cfg.Hosts = []board.Host{board.HostGitHub}
	if cfg.AzureClientID != "" {
		cfg.Hosts = append(cfg.Hosts, board.HostAzure)
	}
	cfg.LocalWrites = true
	return cfg, nil
}

func boardNames(raw []string) []string {
	var names []string
	for _, name := range raw {
		if name = strings.TrimSpace(name); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func hostsFromOldVar() []board.Host {
	raw := os.Getenv("FUDA_SOURCE")
	if raw == "" {
		return []board.Host{board.HostGitHub}
	}
	var hosts []board.Host
	for part := range strings.SplitSeq(raw, ",") {
		hosts = append(hosts, board.Host(strings.TrimSpace(part)))
	}
	return hosts
}

func (c Config) IgnoredVars() []string {
	var ignored []string
	if _, set := os.LookupEnv("FUDA_GITHUB_CLIENT_SECRET"); set {
		ignored = append(ignored, "FUDA_GITHUB_CLIENT_SECRET")
	}
	if !c.Has(board.HostAzure) {
		if c.CookieSecret != "" {
			ignored = append(ignored, "FUDA_COOKIE_SECRET")
		}
		if c.BaseURL != "" {
			ignored = append(ignored, "FUDA_BASE_URL")
		}
	}
	return ignored
}

func (c Config) Has(host board.Host) bool {
	return slices.Contains(c.Hosts, host)
}

func (c Config) validateBoards() error {
	if err := checkParts("FUDA_GITHUB_BOARDS", c.GitHubBoards, "owner/repo"); err != nil {
		return err
	}
	return checkParts("FUDA_AZURE_BOARDS", c.AzureBoards, "org/project/repo")
}

func checkParts(setting string, names []string, form string) error {
	want := strings.Count(form, "/") + 1
	for _, name := range names {
		parts := strings.Split(name, "/")
		if len(parts) != want || slices.Contains(parts, "") {
			return fmt.Errorf("%s: %q is not %s", setting, name, form)
		}
	}
	return nil
}

func (c Config) validate() error {
	if err := c.validateBoards(); err != nil {
		return err
	}
	if c.Has(board.HostLocal) && len(c.Hosts) > 1 {
		return fmt.Errorf("FUDA_HOSTS=local cannot be combined with other hosts")
	}
	for _, host := range c.Hosts {
		switch host {
		case board.HostLocal:
			if c.LocalPath == "" {
				return fmt.Errorf("FUDA_HOSTS=local needs FUDA_LOCAL_PATH")
			}
		case board.HostGitHub:
			if c.GitHubClientID == "" || c.GitHubAppSlug == "" {
				return fmt.Errorf("FUDA_HOSTS=github needs FUDA_GITHUB_CLIENT_ID and FUDA_GITHUB_APP_SLUG")
			}
		case board.HostAzure:
			if c.AzureClientID == "" || c.AzureClientSecret == "" || c.CookieSecret == "" || c.BaseURL == "" {
				return fmt.Errorf("FUDA_HOSTS=azure needs FUDA_AZURE_CLIENT_ID, FUDA_AZURE_CLIENT_SECRET, FUDA_COOKIE_SECRET and FUDA_BASE_URL")
			}
			if len(c.CookieSecret) < minCookieSecretLength {
				return fmt.Errorf("FUDA_COOKIE_SECRET must be at least %d characters", minCookieSecretLength)
			}
		default:
			return fmt.Errorf("unknown host %q: use local, github or azure (github,azure for both)", host)
		}
	}
	return nil
}
