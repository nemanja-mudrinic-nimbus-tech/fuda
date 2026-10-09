package main

import (
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"

	"fuda/internal/app"
)

func openFolder(desktop *application.App, window *application.WebviewWindow, folders *app.Folders) {
	folder, err := desktop.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("Open a folder with Tasks").
		PromptForSingleSelection()
	if err != nil || folder == "" {
		return
	}
	id, err := folders.Add(folder)
	if errors.Is(err, app.ErrNoTasks) {
		desktop.Dialog.Warning().SetTitle("fuda").SetMessage("There is no tasks/ or docs/board/tasks/ folder in that folder.").Show()
		return
	}
	if err != nil {
		desktop.Dialog.Error().SetTitle("fuda").SetMessage("Could not open that folder: " + err.Error()).Show()
		return
	}
	if window != nil {
		window.SetURL(id.Path() + "/")
	}
}
