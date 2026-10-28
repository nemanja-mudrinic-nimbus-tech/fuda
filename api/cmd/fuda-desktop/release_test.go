package main

import (
	"testing"

	"github.com/creativeprojects/go-selfupdate"
)

type fakeRelease struct {
	selfupdate.SourceRelease
	tag        string
	draft      bool
	prerelease bool
}

func (r fakeRelease) GetTagName() string  { return r.tag }
func (r fakeRelease) GetDraft() bool      { return r.draft }
func (r fakeRelease) GetPrerelease() bool { return r.prerelease }

func TestNewestReleasePicksTheHighestPublishedVersion(t *testing.T) {
	releases := []selfupdate.SourceRelease{
		fakeRelease{tag: "v0.9.0"},
		fakeRelease{tag: "v0.10.0"},
		fakeRelease{tag: "v1.0.0", draft: true},
		fakeRelease{tag: "v1.1.0-rc1", prerelease: true},
		fakeRelease{tag: "nightly"},
		fakeRelease{tag: "v0.2.0"},
	}

	got, ok := newestRelease(releases)

	if !ok || got.GetTagName() != "v0.10.0" {
		t.Fatalf("got %v %v, want v0.10.0", got, ok)
	}
}

func TestNewestReleaseFindsNothingWithoutPublishedVersions(t *testing.T) {
	releases := []selfupdate.SourceRelease{fakeRelease{tag: "v1.0.0", draft: true}, fakeRelease{tag: "latest"}}

	if _, ok := newestRelease(releases); ok {
		t.Fatal("expected no release")
	}
	if _, ok := newestRelease(nil); ok {
		t.Fatal("expected no release")
	}
}

func TestIsNewerThanComparesVersions(t *testing.T) {
	cases := []struct {
		tag, current string
		want         bool
	}{
		{"v0.2.0", "0.1.0", true},
		{"v0.10.0", "0.9.0", true},
		{"v0.1.0", "0.1.0", false},
		{"v0.1.0", "0.2.0", false},
		{"v0.1.0", "dev", false},
	}
	for _, c := range cases {
		if got := isNewerThan(c.tag, c.current); got != c.want {
			t.Errorf("isNewerThan(%q, %q) = %v, want %v", c.tag, c.current, got, c.want)
		}
	}
}
