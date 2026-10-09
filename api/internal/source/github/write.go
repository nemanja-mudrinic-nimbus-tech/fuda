package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"fuda/internal/board"
)

func (s *Source) ReadFile(ctx context.Context, branch, path string) ([]byte, string, error) {
	var file struct {
		SHA      string `json:"sha"`
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	err := s.getJSON(ctx, "/repos/"+s.repo+"/contents/"+escapePath(path)+"?ref="+url.QueryEscape(branch), &file)
	if errors.Is(err, errNotFound) {
		return nil, "", board.ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	if file.Encoding != "base64" {
		return nil, "", errors.New("github: unexpected file encoding " + file.Encoding)
	}
	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(file.Content, "\n", ""))
	if err != nil {
		return nil, "", err
	}
	return content, file.SHA, nil
}

func (s *Source) WriteFile(ctx context.Context, branch, path string, content []byte, version, message string) error {
	payload, err := json.Marshal(map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
		"sha":     version,
		"branch":  branch,
	})
	if err != nil {
		return err
	}
	res, err := s.exchange(ctx, http.MethodPut, "/repos/"+s.repo+"/contents/"+escapePath(path), "application/vnd.github+json", "", payload)
	if err != nil {
		return err
	}
	switch {
	case res.status == http.StatusConflict, res.status == http.StatusUnprocessableEntity && bytes.Contains(res.body, []byte("does not match")):
		return board.ErrChanged
	case res.status >= 300:
		return res.failure()
	}
	return nil
}

func (s *Source) LastEditor(ctx context.Context, branch, path string) (string, error) {
	var commits []struct {
		Commit struct {
			Author struct {
				Name string `json:"name"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := s.getJSON(ctx, "/repos/"+s.repo+"/commits?per_page=1&sha="+url.QueryEscape(branch)+"&path="+url.QueryEscape(path), &commits); err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}
	return commits[0].Commit.Author.Name, nil
}

func (s *Source) CanWrite(ctx context.Context) (bool, error) {
	var repo struct {
		Permissions struct {
			Push bool `json:"push"`
		} `json:"permissions"`
	}
	if err := s.getJSON(ctx, "/repos/"+s.repo, &repo); err != nil {
		return false, err
	}
	return repo.Permissions.Push, nil
}
