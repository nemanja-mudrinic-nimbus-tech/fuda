package azure

import (
	"context"
	"net/url"
	"strings"
)

func ListBoards(ctx context.Context) ([]string, error) {
	source := New(Repo{}, "")
	return source.listBoards(ctx)
}

func (s *Source) listBoards(ctx context.Context) ([]string, error) {
	var profile struct {
		ID string `json:"id"`
	}
	if err := s.getJSON(ctx, s.accounts, "/_apis/profile/profiles/me", nil, &profile); err != nil {
		return nil, err
	}
	var accounts struct {
		Value []struct {
			Name string `json:"accountName"`
		} `json:"value"`
	}
	if err := s.getJSON(ctx, s.accounts, "/_apis/accounts", url.Values{"memberId": {profile.ID}}, &accounts); err != nil {
		return nil, err
	}
	var names []string
	for _, account := range accounts.Value {
		var repos struct {
			Value []struct {
				Name    string `json:"name"`
				Project struct {
					Name string `json:"name"`
				} `json:"project"`
			} `json:"value"`
		}
		err := s.getJSON(ctx, s.host+"/"+url.PathEscape(account.Name), "/_apis/git/repositories", nil, &repos)
		if err != nil {
			// an organization that blocks this app must not hide the others
			continue
		}
		for _, repo := range repos.Value {
			if strings.HasPrefix(repo.Name, boardPrefix) {
				names = append(names, account.Name+"/"+repo.Project.Name+"/"+repo.Name)
			}
		}
	}
	return names, nil
}
