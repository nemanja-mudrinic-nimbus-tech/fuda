package github

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

const boardPrefix = "fuda-"

func ListBoards(ctx context.Context, listed []string) ([]string, error) {
	source := &Source{base: apiBase, client: &http.Client{Timeout: 60 * time.Second}}
	return source.listBoards(ctx, listed)
}

func isBoard(name, fullName string, listed []string) bool {
	return strings.HasPrefix(name, boardPrefix) || slices.ContainsFunc(listed, func(l string) bool { return strings.EqualFold(l, fullName) })
}

func (s *Source) listBoards(ctx context.Context, listed []string) ([]string, error) {
	const perPage = 100
	var names []string
	for page := 1; ; page++ {
		var repos []struct {
			FullName string `json:"full_name"`
			Name     string `json:"name"`
		}
		path := "/user/repos?sort=full_name&per_page=" + strconv.Itoa(perPage) + "&page=" + strconv.Itoa(page)
		body, err := s.get(ctx, path, "application/vnd.github+json")
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, err
		}
		for _, r := range repos {
			if isBoard(r.Name, r.FullName, listed) {
				names = append(names, r.FullName)
			}
		}
		if len(repos) < perPage {
			return names, nil
		}
	}
}

func AppInstalled(ctx context.Context) (bool, error) {
	source := &Source{base: apiBase, client: &http.Client{Timeout: 60 * time.Second}}
	return source.appInstalled(ctx)
}

func (s *Source) appInstalled(ctx context.Context) (bool, error) {
	var out struct {
		TotalCount int `json:"total_count"`
	}
	if err := s.getJSON(ctx, "/user/installations?per_page=1", &out); err != nil {
		return false, err
	}
	return out.TotalCount > 0, nil
}
