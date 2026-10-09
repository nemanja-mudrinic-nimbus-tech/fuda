package main

import (
	"cmp"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"fuda/internal/app"
	"fuda/internal/config"
	"fuda/internal/httpapi"
	"fuda/internal/keychain"
	"fuda/internal/login"
	"fuda/internal/web"
)

var (
	version        = "dev"
	githubClientID = ""
	azureClientID  = ""
)

//go:embed icon.png
var icon []byte

//go:embed icon-macos.png
var iconMacOS []byte

var errNoLogin = errors.New("no login is set up")

const noLoginHelp = "No login is set up. Set FUDA_GITHUB_CLIENT_ID or FUDA_AZURE_CLIENT_ID in your local env file and build again with make desktop-macos. Or set one of them in the environment when you start the app."

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(log); err != nil {
		log.Error("fuda stopped", "error", err)
		if errors.Is(err, errNoLogin) {
			showStartError(noLoginHelp)
		}
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	cfg := config.Config{
		SyncCooldown: 30 * time.Second,
		CacheDir:     filepath.Join(cacheRoot, "fuda"),
	}

	var desktop *application.App
	var logins []httpapi.Login
	for _, hostConfig := range []login.DeviceConfig{
		{
			Host:     "github",
			ClientID: cmp.Or(os.Getenv("FUDA_GITHUB_CLIENT_ID"), githubClientID),
		},
		{
			Host:     "azure",
			ClientID: cmp.Or(os.Getenv("FUDA_AZURE_CLIENT_ID"), azureClientID),
			Tenant:   cmp.Or(os.Getenv("FUDA_AZURE_TENANT"), "organizations"),
		},
	} {
		if hostConfig.ClientID == "" {
			continue
		}
		hostConfig.Store = keychain.New(hostConfig.Host)
		hostConfig.Open = func(url string) error { return desktop.Browser.OpenURL(url) }
		logins = append(logins, login.NewDevice(log, hostConfig))
		cfg.Sources = append(cfg.Sources, config.Source(hostConfig.Host))
	}
	if len(logins) == 0 {
		return errNoLogin
	}
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	folders, err := app.NewFolders(filepath.Join(configRoot, "fuda", "folders.json"))
	if err != nil {
		return err
	}
	desktop = application.New(application.Options{
		Name: "fuda",
		Icon: appIcon(),
		Assets: application.AssetOptions{
			Handler: httpapi.NewHandler(log, app.NewBoards(log, cfg, folders), web.Dist(), logins...),
		},
	})

	menu := desktop.NewMenu()
	menu.AddRole(application.AppMenu)
	var window *application.WebviewWindow
	file := menu.AddSubmenu("File")
	file.Add("Open folder…").SetAccelerator("CmdOrCtrl+O").OnClick(func(*application.Context) {
		openFolder(desktop, window, folders)
	})
	menu.AddRole(application.EditMenu)
	menu.AddRole(application.WindowMenu)
	help := menu.AddSubmenu("Help")
	help.Add("Check for updates…").OnClick(func(*application.Context) {
		go checkForUpdates(desktop, log, true)
	})
	desktop.Menu.Set(menu)

	window = desktop.Window.NewWithOptions(application.WebviewWindowOptions{
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

func appIcon() []byte {
	if runtime.GOOS == "darwin" {
		return iconMacOS
	}
	return icon
}
