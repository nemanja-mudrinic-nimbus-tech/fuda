package main

import (
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"

	"fuda/internal/workspace"
)

func newMenu(desktop *application.App, folders *workspace.Folders, log *slog.Logger) *application.Menu {
	menu := desktop.NewMenu()
	menu.AddRole(application.AppMenu)
	file := menu.AddSubmenu("File")
	file.Add("Open folder…").SetAccelerator("CmdOrCtrl+O").OnClick(func(*application.Context) {
		openFolder(desktop, folders)
	})
	menu.AddRole(application.EditMenu)
	menu.AddRole(application.WindowMenu)
	help := menu.AddSubmenu("Help")
	help.Add("Check for updates…").OnClick(func(*application.Context) {
		go checkForUpdates(desktop, log, true)
	})
	return menu
}
