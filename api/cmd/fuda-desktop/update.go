package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const releasesRepo = "nemanja-mudrinic-nimbus-tech/fuda"

func checkForUpdates(desktop *application.App, log *slog.Logger, userRequested bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	latest, found, err := findNewestRelease(ctx)
	switch {
	case err != nil:
		log.Warn("check for updates", "error", err)
		if userRequested {
			desktop.Dialog.Warning().SetTitle("fuda").SetMessage("Could not check for updates. Are you online?").Show()
		}
		return
	case !found || !isNewerThan(latest.GetTagName(), version):
		if userRequested {
			desktop.Dialog.Info().SetTitle("fuda").SetMessage("You have the newest version.").Show()
		}
		return
	}

	question := desktop.Dialog.Question().SetTitle("fuda").
		SetMessage(fmt.Sprintf("fuda %s is ready. You have %s. Open the download page?", strings.TrimPrefix(latest.GetTagName(), "v"), version))
	question.AddButton("Open page").OnClick(func() {
		if err := desktop.Browser.OpenURL(latest.GetURL()); err != nil {
			log.Error("open the release page", "error", err)
		}
	})
	question.AddButton("Not now").SetAsCancel()
	question.Show()
}

func findNewestRelease(ctx context.Context) (selfupdate.SourceRelease, bool, error) {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return nil, false, err
	}
	releases, err := source.ListReleases(ctx, selfupdate.ParseSlug(releasesRepo))
	if err != nil {
		return nil, false, err
	}
	latest, found := newestRelease(releases)
	return latest, found, nil
}
