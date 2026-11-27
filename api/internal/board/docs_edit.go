package board

import (
	"context"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
)

var ErrNotEditable = errors.New("this document cannot be edited")

var boardConfigFiles = []string{"stages.md", "labels.md", "people.md", "repos.md"}

type DocSource struct {
	Content string `json:"content"`
	Version string `json:"version"`
}

type SaveDocRequest struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Content string `json:"content"`
	Message string `json:"message"`
}

func (s *Service) editableDoc(relative string) (string, error) {
	repoPath := path.Clean(path.Join(s.opts.DocsRoot, relative))
	if !strings.HasPrefix(repoPath, s.opts.DocsRoot+"/") {
		return "", ErrNotEditable
	}
	for _, folder := range []string{"tasks", "archive"} {
		if strings.HasPrefix(repoPath, path.Join(s.opts.BoardDir, folder)+"/") {
			return "", ErrNotEditable
		}
	}
	if path.Dir(repoPath) == s.opts.BoardDir && slices.Contains(boardConfigFiles, path.Base(repoPath)) {
		return "", ErrNotEditable
	}
	snap := s.snapshot.Load()
	if snap == nil {
		return "", ErrNotFound
	}
	if _, ok := snap.docs[repoPath]; !ok {
		return "", ErrNotFound
	}
	return repoPath, nil
}

func (s *Service) DocSource(ctx context.Context, relative string) (DocSource, error) {
	editor, err := s.requireEditor(ctx)
	if err != nil {
		return DocSource{}, err
	}
	repoPath, err := s.editableDoc(relative)
	if err != nil {
		return DocSource{}, err
	}
	content, version, err := editor.ReadFile(ctx, s.opts.WorkBranch, repoPath)
	if err != nil {
		return DocSource{}, err
	}
	return DocSource{Content: string(content), Version: version}, nil
}

func (s *Service) SaveDoc(ctx context.Context, req SaveDocRequest) error {
	editor, err := s.requireEditor(ctx)
	if err != nil {
		return err
	}
	repoPath, err := s.editableDoc(req.Path)
	if err != nil {
		return err
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = "docs: edit " + strings.TrimPrefix(repoPath, s.opts.DocsRoot+"/")
	}
	err = editor.WriteFile(ctx, s.opts.WorkBranch, repoPath, []byte(req.Content), req.Version, message)
	if errors.Is(err, ErrStale) {
		who := s.lastEditor(ctx, editor, repoPath)
		s.syncAfterWrite(ctx)
		return &ConflictError{Message: fmt.Sprintf("%s changed this document since you started editing. Copy your changes, then reload the page.", who)}
	}
	if err != nil {
		return err
	}
	s.syncAfterWrite(ctx)
	return nil
}
