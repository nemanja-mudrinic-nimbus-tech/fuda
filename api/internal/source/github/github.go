package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"fuda/internal/board"
)

const apiBase = "https://api.github.com"

var errNotFound = errors.New("github: not found")

type Source struct {
	repo     string
	docsRoot string
	base     string
	client   *http.Client

	mu    sync.Mutex
	heads map[string]cachedHead
}

type cachedHead struct{ etag, sha string }

func New(repo, docsRoot string) *Source {
	return &Source{
		repo:     repo,
		docsRoot: docsRoot,
		base:     apiBase,
		client:   &http.Client{Timeout: 60 * time.Second},
		heads:    map[string]cachedHead{},
	}
}

func (s *Source) CodeURL(branch string) func(string) string {
	return func(repoPath string) string {
		return "https://github.com/" + s.repo + "/blob/" + branch + "/" + repoPath
	}
}

func (s *Source) PRLink() string {
	return "https://github.com/" + s.repo + "/pull/{n}"
}

func (s *Source) Head(ctx context.Context, branch string) (string, error) {
	s.mu.Lock()
	cached := s.heads[branch]
	s.mu.Unlock()

	res, err := s.send(ctx, "/repos/"+s.repo+"/branches/"+url.PathEscape(branch), "application/vnd.github+json", cached.etag)
	if err != nil {
		return "", err
	}
	switch {
	case res.status == http.StatusNotModified:
		return cached.sha, nil
	case res.status == http.StatusNotFound:
		return "", s.missingBranch(ctx)
	case res.status >= 300:
		return "", res.failure()
	}
	var out struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(res.body, &out); err != nil {
		return "", err
	}
	s.mu.Lock()
	s.heads[branch] = cachedHead{etag: res.etag, sha: out.Commit.SHA}
	s.mu.Unlock()
	return out.Commit.SHA, nil
}

func (s *Source) missingBranch(ctx context.Context) error {
	_, err := s.get(ctx, "/repos/"+s.repo, "application/vnd.github+json")
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound):
		return board.ErrBranchMissing
	}
	return err
}

func (s *Source) Files(ctx context.Context, branch string) (map[string][]byte, error) {
	body, err := s.get(ctx, "/repos/"+s.repo+"/zipball/"+url.PathEscape(branch), "")
	if errors.Is(err, errNotFound) {
		if err := s.missingBranch(ctx); err != nil {
			return nil, err
		}
		return map[string][]byte{}, nil
	}
	if err != nil {
		return nil, err
	}
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("zipball: %w", err)
	}
	files := map[string][]byte{}
	for _, f := range archive.File {
		_, rel, found := strings.Cut(f.Name, "/")
		if !found || f.FileInfo().IsDir() || !strings.HasPrefix(rel, s.docsRoot+"/") || !board.WantedFile(rel, int64(f.UncompressedSize64)) {
			continue
		}
		content, err := readZipFile(f)
		if err != nil {
			return nil, err
		}
		files[rel] = content
	}
	return files, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

func (s *Source) OpenPRs(ctx context.Context, repo string) ([]board.PullRequest, error) {
	var pulls []struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
		Title   string `json:"title"`
		Head    struct {
			Ref string `json:"ref"`
		} `json:"head"`
	}
	err := s.getJSON(ctx, "/repos/"+escapePath(repo)+"/pulls?state=open&per_page=100", &pulls)
	if errors.Is(err, errNotFound) {
		return nil, board.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	out := make([]board.PullRequest, len(pulls))
	for i, p := range pulls {
		out[i] = board.PullRequest{Number: p.Number, URL: p.HTMLURL, Title: p.Title, Branch: p.Head.Ref}
	}
	return out, nil
}

func escapePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func (s *Source) getJSON(ctx context.Context, path string, v any) error {
	body, err := s.get(ctx, path, "application/vnd.github+json")
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

type response struct {
	path   string
	status int
	body   []byte
	etag   string
}

func (r response) failure() error {
	return fmt.Errorf("github %s: status %d: %s", r.path, r.status, truncate(r.body))
}

func (s *Source) get(ctx context.Context, path, accept string) ([]byte, error) {
	res, err := s.send(ctx, path, accept, "")
	if err != nil {
		return nil, err
	}
	switch {
	case res.status == http.StatusNotFound:
		return nil, errNotFound
	case res.status >= 300:
		return nil, res.failure()
	}
	return res.body, nil
}

func (s *Source) send(ctx context.Context, path, accept, ifNoneMatch string) (response, error) {
	return s.exchange(ctx, http.MethodGet, path, accept, ifNoneMatch, nil)
}

func (s *Source) exchange(ctx context.Context, method, path, accept, ifNoneMatch string, payload []byte) (response, error) {
	token := board.TokenFrom(ctx)
	if token == "" {
		return response{}, board.ErrUnauthorized
	}
	req, err := http.NewRequestWithContext(ctx, method, s.base+path, bytes.NewReader(payload))
	if err != nil {
		return response{}, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	res, err := s.client.Do(req)
	if err != nil {
		return response{}, err
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return response{}, err
	}
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return response{}, board.ErrUnauthorized
	case http.StatusForbidden:
		return response{}, board.ErrForbidden
	}
	return response{path: path, status: res.StatusCode, body: body, etag: res.Header.Get("ETag")}, nil
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}
