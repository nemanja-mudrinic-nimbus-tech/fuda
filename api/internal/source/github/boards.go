package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const boardPrefix = "fuda-"

func ListBoards(ctx context.Context) ([]string, error) {
	source := &Source{base: apiBase, client: &http.Client{Timeout: 60 * time.Second}}
	return source.listBoards(ctx)
}

func (s *Source) listBoards(ctx context.Context) ([]string, error) {
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
			if strings.HasPrefix(r.Name, boardPrefix) {
				names = append(names, r.FullName)
			}
		}
		if len(repos) < perPage {
			return names, nil
		}
	}
}
