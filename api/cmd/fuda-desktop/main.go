package main

import (
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"

	"fuda/internal/board"
	"fuda/internal/config"
	"fuda/internal/httpapi"
	"fuda/internal/keychain"
	"fuda/internal/login"
	"fuda/internal/web"
	"fuda/internal/workspace"
)

var (
	version        = "dev"
	githubClientID = ""
	githubAppSlug  = ""
	azureClientID  = ""
	githubBoards   = ""
	azureBoards    = ""
)

//go:embed icon.png
var icon []byte

//go:embed icon-macos.png
var iconMacOS []byte

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(log); err != nil {
		log.Error("fuda stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	cfg, err := config.LoadDesktop(config.DesktopDefaults{
		GitHubClientID: githubClientID,
		GitHubAppSlug:  githubAppSlug,
		AzureClientID:  azureClientID,
		GitHubBoards:   githubBoards,
		AzureBoards:    azureBoards,
		CacheDir:       filepath.Join(cacheRoot, "fuda"),
	})
	if err != nil {
		return err
	}

	var desktop *application.App
	openBrowser := func(url string) error { return desktop.Browser.OpenURL(url) }
	logins := newLogins(log, cfg, openBrowser)

	configRoot, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	folders, err := workspace.NewFolders(filepath.Join(configRoot, "fuda", "folders.json"))
	if err != nil {
		return err
	}
	desktop = application.New(application.Options{
		Name: "fuda",
		Icon: appIcon(),
		Assets: application.AssetOptions{
			Handler: httpapi.NewHandler(log, workspace.NewRegistry(log, cfg, folders), web.Dist(), logins...),
		},
	})
	desktop.Menu.Set(newMenu(desktop, folders, log))
	desktop.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "fuda",
		Width:     1280,
		Height:    820,
		MinWidth:  720,
		MinHeight: 480,
		URL:       "/",
	})
	go checkForUpdates(desktop, log, false)
	return desktop.Run()
}

func newLogins(log *slog.Logger, cfg config.Config, open func(url string) error) []httpapi.Login {
	var logins []httpapi.Login
	for _, host := range cfg.Hosts {
		device := login.DeviceConfig{Host: host, Store: keychain.New(string(host)), Open: open}
		switch host {
		case board.HostGitHub:
			device.ClientID = cfg.GitHubClientID
		case board.HostAzure:
			device.ClientID = cfg.AzureClientID
			device.Tenant = cfg.AzureTenant
		}
		logins = append(logins, login.NewKeychainDevice(log, device))
	}
	return logins
}

func appIcon() []byte {
	if runtime.GOOS == "darwin" {
		return iconMacOS
	}
	return icon
}
