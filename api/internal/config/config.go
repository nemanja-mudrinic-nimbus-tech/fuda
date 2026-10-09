package config

import (
	"fmt"
	"slices"
	"time"

	"github.com/caarlos0/env/v11"
)

type Source string

const (
	SourceLocal  Source = "local"
	SourceGitHub Source = "github"
	SourceAzure  Source = "azure"
)

type Config struct {
	Addr  string `env:"FUDA_ADDR" envDefault:":8080"`
	Title string `env:"FUDA_TITLE"`

	BaseURL      string `env:"FUDA_BASE_URL"`
	CookieSecret string `env:"FUDA_COOKIE_SECRET"`

	Sources   []Source `env:"FUDA_SOURCE" envDefault:"github"`
	LocalPath string   `env:"FUDA_LOCAL_PATH"`

	GitHubClientID            string `env:"FUDA_GITHUB_CLIENT_ID"`
	IgnoredGitHubClientSecret string `env:"FUDA_GITHUB_CLIENT_SECRET"`

	AzureTenant       string `env:"FUDA_AZURE_TENANT" envDefault:"organizations"`
	AzureClientID     string `env:"FUDA_AZURE_CLIENT_ID"`
	AzureClientSecret string `env:"FUDA_AZURE_CLIENT_SECRET"`

	WatchMain    bool          `env:"FUDA_WATCH_MAIN" envDefault:"false"`
	SyncCooldown time.Duration `env:"FUDA_SYNC_COOLDOWN" envDefault:"30s"`
	CacheDir     string        `env:"FUDA_CACHE_DIR" envDefault:"./.fuda-cache"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	return cfg, cfg.validate()
}

func (c Config) IgnoredVars() []string {
	var ignored []string
	if c.IgnoredGitHubClientSecret != "" {
		ignored = append(ignored, "FUDA_GITHUB_CLIENT_SECRET")
	}
	if !c.Has(SourceAzure) {
		if c.CookieSecret != "" {
			ignored = append(ignored, "FUDA_COOKIE_SECRET")
		}
		if c.BaseURL != "" {
			ignored = append(ignored, "FUDA_BASE_URL")
		}
	}
	return ignored
}

func (c Config) Has(source Source) bool {
	return slices.Contains(c.Sources, source)
}

func (c Config) validate() error {
	if c.Has(SourceLocal) && len(c.Sources) > 1 {
		return fmt.Errorf("FUDA_SOURCE=local cannot be combined with other sources")
	}
	for _, source := range c.Sources {
		switch source {
		case SourceLocal:
			if c.LocalPath == "" {
				return fmt.Errorf("FUDA_SOURCE=local needs FUDA_LOCAL_PATH")
			}
		case SourceGitHub:
			if c.GitHubClientID == "" {
				return fmt.Errorf("FUDA_SOURCE=github needs FUDA_GITHUB_CLIENT_ID")
			}
		case SourceAzure:
			if c.AzureClientID == "" || c.AzureClientSecret == "" || c.CookieSecret == "" || c.BaseURL == "" {
				return fmt.Errorf("FUDA_SOURCE=azure needs FUDA_AZURE_CLIENT_ID, FUDA_AZURE_CLIENT_SECRET, FUDA_COOKIE_SECRET and FUDA_BASE_URL")
			}
		default:
			return fmt.Errorf("unknown FUDA_SOURCE %q: use local, github or azure (github,azure for both)", source)
		}
	}
	return nil
}
