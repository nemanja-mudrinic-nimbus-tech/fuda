package azure

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
	"time"

	"fuda/internal/board"
)

const apiVersion = "7.1"

var errNotFound = errors.New("azure: not found")

type Repo struct {
	Org     string
	Project string
	Name    string
}

const (
	hostBase     = "https://dev.azure.com"
	accountsBase = "https://app.vssps.visualstudio.com"
	boardPrefix  = "fuda-"
)

func (r Repo) WebURL() string {
	return hostBase + "/" + url.PathEscape(r.Org) + "/" + url.PathEscape(r.Project) + "/_git/" + url.PathEscape(r.Name)
}

func (r Repo) apiBase(host string) string {
	return host + "/" + url.PathEscape(r.Org) + "/" + url.PathEscape(r.Project) + "/_apis/git/repositories/" + url.PathEscape(r.Name)
}

type Source struct {
	repo     Repo
	docsRoot string
	host     string
	accounts string
	client   *http.Client
}

func New(repo Repo, docsRoot string) *Source {
	return &Source{
		repo:     repo,
		docsRoot: docsRoot,
		host:     hostBase,
		accounts: accountsBase,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *Source) base() string { return s.repo.apiBase(s.host) }

func (s *Source) CodeURL(branch string) func(string) string {
	return func(repoPath string) string {
		return s.repo.WebURL() + "?path=" + url.QueryEscape("/"+repoPath) + "&version=GB" + url.QueryEscape(branch)
	}
}

func (s *Source) PRLink() string {
	return s.repo.WebURL() + "/pullrequest/{n}"
}

func (s *Source) Head(ctx context.Context, branch string) (string, error) {
	var out struct {
		Value []struct {
			Name     string `json:"name"`
			ObjectID string `json:"objectId"`
		} `json:"value"`
	}
	if err := s.getJSON(ctx, s.base(), "/refs", url.Values{"filter": {"heads/" + branch}}, &out); err != nil {
		return "", err
	}
	for _, ref := range out.Value {
		if ref.Name == "refs/heads/"+branch {
			return ref.ObjectID, nil
		}
	}
	return "", board.ErrBranchMissing
}

func (s *Source) Files(ctx context.Context, branch string) (map[string][]byte, error) {
	res, err := s.send(ctx, http.MethodGet, s.base()+"/items", url.Values{
		"path":                          {"/" + s.docsRoot},
		"$format":                       {"zip"},
		"download":                      {"true"},
		"versionDescriptor.version":     {branch},
		"versionDescriptor.versionType": {"branch"},
	}, nil)
	if err != nil {
		return nil, err
	}
	if res.status == http.StatusNotFound {
		return map[string][]byte{}, nil
	}
	if res.status >= 300 {
		return nil, res.failure()
	}
	body := res.body
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("items zip: %w", err)
	}
	files := map[string][]byte{}
	for _, f := range archive.File {
		name := strings.TrimPrefix(f.Name, "/")
		if f.FileInfo().IsDir() || !strings.HasPrefix(name, s.docsRoot+"/") || !board.WantedFile(name, int64(f.UncompressedSize64)) {
			continue
		}
		content, err := readZipFile(f)
		if err != nil {
			return nil, err
		}
		files[name] = content
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
	parts := strings.Split(repo, "/")
	if len(parts) != 3 {
		return nil, board.ErrNotFound
	}
	code := Repo{Org: parts[0], Project: parts[1], Name: parts[2]}
	var pulls struct {
		Value []struct {
			ID            int    `json:"pullRequestId"`
			Title         string `json:"title"`
			SourceRefName string `json:"sourceRefName"`
		} `json:"value"`
	}
	err := s.getJSON(ctx, code.apiBase(s.host), "/pullrequests", url.Values{
		"searchCriteria.status": {"active"},
		"$top":                  {"100"},
	}, &pulls)
	if errors.Is(err, errNotFound) {
		return nil, board.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	out := make([]board.PullRequest, len(pulls.Value))
	for i, p := range pulls.Value {
		out[i] = board.PullRequest{
			Number: p.ID,
			URL:    fmt.Sprintf("%s/pullrequest/%d", code.WebURL(), p.ID),
			Title:  p.Title,
			Branch: strings.TrimPrefix(p.SourceRefName, "refs/heads/"),
		}
	}
	return out, nil
}

type response struct {
	url    string
	status int
	body   []byte
}

func (r response) failure() error {
	return fmt.Errorf("azure %s: status %d: %s", r.url, r.status, truncate(r.body))
}

func (s *Source) getJSON(ctx context.Context, base, path string, query url.Values, v any) error {
	res, err := s.send(ctx, http.MethodGet, base+path, query, nil)
	if err != nil {
		return err
	}
	switch {
	case res.status == http.StatusNotFound:
		return errNotFound
	case res.status >= 300:
		return res.failure()
	}
	return json.Unmarshal(res.body, v)
}

func (s *Source) send(ctx context.Context, method, target string, query url.Values, payload []byte) (response, error) {
	token := board.TokenFrom(ctx)
	if token == "" {
		return response{}, board.ErrUnauthorized
	}
	if query == nil {
		query = url.Values{}
	}
	query.Set("api-version", apiVersion)
	req, err := http.NewRequestWithContext(ctx, method, target+"?"+query.Encode(), bytes.NewReader(payload))
	if err != nil {
		return response{}, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
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
	case http.StatusNonAuthoritativeInfo, http.StatusUnauthorized:
		return response{}, board.ErrUnauthorized
	case http.StatusForbidden:
		return response{}, board.ErrForbidden
	}
	return response{url: target, status: res.StatusCode, body: body}, nil
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}
