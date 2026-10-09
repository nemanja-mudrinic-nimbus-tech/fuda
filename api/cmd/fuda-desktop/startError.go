package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func showStartError(message string) {
	desktop := application.New(application.Options{Name: "fuda", Icon: appIcon()})
	desktop.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		desktop.Dialog.Error().SetTitle("fuda").SetMessage(message).Show()
		desktop.Quit()
	})
	_ = desktop.Run()
}
