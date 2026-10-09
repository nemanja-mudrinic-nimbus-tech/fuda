package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const releasesRepo = "nemanja-mudrinic-nimbus-tech/fuda"

func checkForUpdates(desktop *application.App, log *slog.Logger, userRequested bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(releasesRepo))
	switch {
	case err != nil:
		log.Warn("check for updates", "error", err)
		if userRequested {
			desktop.Dialog.Warning().SetTitle("fuda").SetMessage("Could not check for updates. Are you online?").Show()
		}
		return
	case !found || version == "dev" || !latest.GreaterThan(version):
		if userRequested {
			desktop.Dialog.Info().SetTitle("fuda").SetMessage("You have the newest version.").Show()
		}
		return
	}

	question := desktop.Dialog.Question().SetTitle("fuda").
		SetMessage(fmt.Sprintf("fuda %s is ready. You have %s. Update now?", latest.Version(), version))
	question.AddButton("Update").OnClick(func() { install(desktop, log, latest) })
	question.AddButton("Not now").SetAsCancel()
	question.Show()
}

func install(desktop *application.App, log *slog.Logger, latest *selfupdate.Release) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	exe, err := os.Executable()
	if err == nil {
		err = selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe)
	}
	if err != nil {
		log.Error("update fuda", "error", err)
		desktop.Dialog.Warning().SetTitle("fuda").SetMessage("The update failed. Try again later.").Show()
		return
	}
	desktop.Dialog.Info().SetTitle("fuda").SetMessage("Updated. Close fuda and open it again to use " + latest.Version() + ".").Show()
}
