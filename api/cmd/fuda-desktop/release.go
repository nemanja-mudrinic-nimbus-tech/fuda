package main

import (
	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
)

func newestRelease(releases []selfupdate.SourceRelease) (selfupdate.SourceRelease, bool) {
	var newest selfupdate.SourceRelease
	var newestVersion *semver.Version
	for _, release := range releases {
		if release.GetDraft() || release.GetPrerelease() {
			continue
		}
		v, err := semver.NewVersion(release.GetTagName())
		if err != nil {
			continue
		}
		if newestVersion == nil || v.GreaterThan(newestVersion) {
			newest, newestVersion = release, v
		}
	}
	return newest, newest != nil
}

func isNewerThan(tag, current string) bool {
	latest, err := semver.NewVersion(tag)
	if err != nil {
		return false
	}
	running, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	return latest.GreaterThan(running)
}
