package main

import (
	"context"
	"fmt"
	"log/slog"
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
		SetMessage(fmt.Sprintf("fuda %s is ready. You have %s. Open the download page?", latest.Version(), version))
	question.AddButton("Open page").OnClick(func() {
		if err := desktop.Browser.OpenURL(latest.URL); err != nil {
			log.Error("open the release page", "error", err)
		}
	})
	question.AddButton("Not now").SetAsCancel()
	question.Show()
}
