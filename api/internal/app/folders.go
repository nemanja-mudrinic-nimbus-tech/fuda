package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"fuda/internal/board"
)

var ErrNoTasks = errors.New("no tasks/ or docs/board/tasks/ in that folder")

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._ -]+`)

type Folders struct {
	storePath string

	mu    sync.Mutex
	paths []string
}

func NewFolders(storePath string) (*Folders, error) {
	f := &Folders{storePath: storePath}
	if storePath == "" {
		return f, nil
	}
	raw, err := os.ReadFile(storePath)
	if errors.Is(err, os.ErrNotExist) {
		return f, nil
	}
	if err != nil {
		return nil, err
	}
	var remembered []string
	if err := json.Unmarshal(raw, &remembered); err != nil {
		return nil, fmt.Errorf("%s: %w", storePath, err)
	}
	for _, p := range remembered {
		if hasTasks(p) {
			f.paths = append(f.paths, p)
		}
	}
	return f, nil
}

func (f *Folders) Add(folder string) (board.BoardID, error) {
	abs, err := filepath.Abs(folder)
	if err != nil {
		return board.BoardID{}, err
	}
	if !hasTasks(abs) {
		return board.BoardID{}, ErrNoTasks
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !slices.Contains(f.paths, abs) {
		f.paths = append(f.paths, abs)
		if err := f.save(); err != nil {
			f.paths = f.paths[:len(f.paths)-1]
			return board.BoardID{}, err
		}
	}
	return f.ids()[slices.Index(f.paths, abs)], nil
}

func (f *Folders) List() []board.BoardID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ids()
}

func (f *Folders) Path(id board.BoardID) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := slices.Index(f.ids(), id)
	if i < 0 {
		return "", false
	}
	return f.paths[i], true
}

// Two folders with one name get -2, -3 in the order they were added.
func (f *Folders) ids() []board.BoardID {
	taken := map[string]bool{}
	ids := make([]board.BoardID, len(f.paths))
	for i, p := range f.paths {
		base := folderName(p)
		name := base
		for n := 2; taken[name]; n++ {
			name = base + "-" + strconv.Itoa(n)
		}
		taken[name] = true
		ids[i] = board.BoardID{Host: "local", Repo: name}
	}
	return ids
}

func (f *Folders) save() error {
	if f.storePath == "" {
		return nil
	}
	raw, err := json.Marshal(f.paths)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.storePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.storePath, raw, 0o600)
}

func folderName(folder string) string {
	name := strings.TrimLeft(unsafeName.ReplaceAllString(filepath.Base(folder), "-"), "._ -")
	if name == "" {
		return "folder"
	}
	return name
}

func hasTasks(folder string) bool {
	for _, tasks := range []string{filepath.Join(boardDir, "tasks"), "tasks"} {
		if info, err := os.Stat(filepath.Join(folder, tasks)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}
