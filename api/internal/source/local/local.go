package local

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"fuda/internal/board"
)

type Source struct {
	root       string
	docsRoot   string
	boardDir   string
	workBranch string
	tasksRepo  bool
}

func New(root, docsRoot, boardDir, workBranch string) *Source {
	s := &Source{root: root, docsRoot: docsRoot, boardDir: boardDir, workBranch: workBranch}
	s.tasksRepo = !isDir(filepath.Join(root, filepath.FromSlash(boardDir), "tasks")) && isDir(filepath.Join(root, "tasks"))
	return s
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func (s *Source) Head(ctx context.Context, branch string) (string, error) {
	if branch == s.workBranch {
		return s.workingTreeHash()
	}
	if s.tasksRepo {
		return "", board.ErrBranchMissing
	}
	ref, err := s.resolve(ctx, branch)
	if err != nil {
		return "", err
	}
	out, err := s.git(ctx, "rev-parse", ref)
	return strings.TrimSpace(string(out)), err
}

func (s *Source) Files(ctx context.Context, branch string) (map[string][]byte, error) {
	if branch == s.workBranch {
		return s.workingTreeFiles()
	}
	if s.tasksRepo {
		return nil, board.ErrBranchMissing
	}
	ref, err := s.resolve(ctx, branch)
	if err != nil {
		return nil, err
	}
	if _, err := s.git(ctx, "rev-parse", "--verify", "--quiet", ref+":"+s.docsRoot); err != nil {
		return map[string][]byte{}, nil
	}
	archive, err := s.git(ctx, "archive", "--format=tar", ref, s.docsRoot)
	if err != nil {
		return nil, err
	}
	return docsFromTar(archive)
}

func (s *Source) resolve(ctx context.Context, branch string) (string, error) {
	for _, ref := range []string{"origin/" + branch, branch} {
		if _, err := s.git(ctx, "rev-parse", "--verify", "--quiet", ref+"^{commit}"); err == nil {
			return ref, nil
		}
	}
	return "", board.ErrBranchMissing
}

func (s *Source) git(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", s.root}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func (s *Source) walkDocs(visit func(rel string, d fs.DirEntry) error) error {
	dir := filepath.Join(s.root, s.docsRoot)
	if s.tasksRepo {
		dir = s.root
	}
	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			if s.tasksRepo && p != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil || !board.WantedFile(d.Name(), info.Size()) {
			return err
		}
		rel, err := filepath.Rel(s.root, p)
		if err != nil {
			return err
		}
		return visit(s.boardPath(filepath.ToSlash(rel)), d)
	})
}

// A tasks repository keeps tasks/ at its root; the board reads every board as if it sat in boardDir.
func (s *Source) boardPath(rel string) string {
	if s.tasksRepo {
		return path.Join(s.boardDir, rel)
	}
	return rel
}

func (s *Source) diskPath(boardPath string) (string, error) {
	rel := path.Clean(boardPath)
	if s.tasksRepo {
		var ok bool
		if rel, ok = strings.CutPrefix(rel, s.boardDir+"/"); !ok {
			return "", board.ErrNotFound
		}
	}
	if !fs.ValidPath(rel) {
		return "", board.ErrNotFound
	}
	return filepath.Join(s.root, filepath.FromSlash(rel)), nil
}

func (s *Source) workingTreeHash() (string, error) {
	h := sha256.New()
	err := s.walkDocs(func(rel string, d fs.DirEntry) error {
		info, err := d.Info()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(h, "%s\x00%d\x00%d\n", rel, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	return "worktree-" + hex.EncodeToString(h.Sum(nil))[:12], err
}

func (s *Source) workingTreeFiles() (map[string][]byte, error) {
	files := map[string][]byte{}
	err := s.walkDocs(func(rel string, _ fs.DirEntry) error {
		disk, err := s.diskPath(rel)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(disk)
		files[rel] = content
		return err
	})
	return files, err
}

func docsFromTar(archive []byte) (map[string][]byte, error) {
	files := map[string][]byte{}
	r := tar.NewReader(bytes.NewReader(archive))
	for {
		h, err := r.Next()
		if errors.Is(err, io.EOF) {
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag != tar.TypeReg || !board.WantedFile(h.Name, h.Size) {
			continue
		}
		content, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		files[h.Name] = content
	}
}
