package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"fuda/internal/board"
)

func (s *Source) ReadFile(_ context.Context, _, boardPath string) ([]byte, string, error) {
	disk, err := s.diskPath(boardPath)
	if err != nil {
		return nil, "", err
	}
	content, err := os.ReadFile(disk)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", board.ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	return content, versionOf(content), nil
}

func (s *Source) WriteFile(_ context.Context, _, boardPath string, content []byte, version, _ string) error {
	disk, err := s.diskPath(boardPath)
	if err != nil {
		return err
	}
	current, err := os.ReadFile(disk)
	if errors.Is(err, fs.ErrNotExist) {
		return board.ErrChanged
	}
	if err != nil {
		return err
	}
	if versionOf(current) != version {
		return board.ErrChanged
	}
	info, err := os.Stat(disk)
	if err != nil {
		return err
	}
	return replaceFile(disk, content, info.Mode().Perm())
}

func (s *Source) LastEditor(context.Context, string, string) (string, error) {
	return "", nil
}

func versionOf(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func replaceFile(disk string, content []byte, mode fs.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(disk), ".fuda-write-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), disk)
}

func (s *Source) CanWrite(context.Context) (bool, error) {
	dir, err := s.diskPath(path.Join(s.boardDir, "tasks", "x"))
	if err != nil {
		return false, err
	}
	probe, err := os.CreateTemp(filepath.Dir(dir), ".fuda-probe-*")
	if err != nil {
		return false, nil
	}
	_ = probe.Close()
	_ = os.Remove(probe.Name())
	return true, nil
}
