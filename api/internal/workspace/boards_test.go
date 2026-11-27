package workspace

import (
	"testing"

	"fuda/internal/board"
)

func TestOnlyBoardsInANonFudaRepositoryAreInACodeRepo(t *testing.T) {
	tests := []struct {
		id   board.ID
		want bool
	}{
		{board.ID{Host: board.HostGitHub, Repo: "o/fuda-tickets"}, false},
		{board.ID{Host: board.HostGitHub, Repo: "o/fuda"}, true},
		{board.ID{Host: board.HostGitHub, Repo: "o/app"}, true},
		{board.ID{Host: board.HostAzure, Repo: "org/project/fuda-tickets"}, false},
		{board.ID{Host: board.HostAzure, Repo: "org/project/app"}, true},
		{board.ID{Host: board.HostLocal, Repo: "app"}, false},
	}
	for _, tt := range tests {
		if got := inCodeRepo(tt.id); got != tt.want {
			t.Errorf("%+v: got %v, want %v", tt.id, got, tt.want)
		}
	}
}
