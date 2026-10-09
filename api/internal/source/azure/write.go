package azure

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"fuda/internal/board"
)

var errPushRefused = errors.New("azure: the push was refused")

func readOnlyOn(err error) error {
	if errors.Is(err, board.ErrForbidden) || errors.Is(err, errNotFound) {
		return nil
	}
	return err
}

const (
	gitNamespace     = "2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87"
	contributeBit    = "4"
	pushAttempts     = 3
	pushBranchPrefix = "refs/heads/"
)

type item struct {
	ObjectID string `json:"objectId"`
	Content  string `json:"content"`
}

func (s *Source) item(ctx context.Context, branch, path string, withContent bool) (item, error) {
	query := url.Values{
		"path":                          {"/" + path},
		"$format":                       {"json"},
		"versionDescriptor.version":     {branch},
		"versionDescriptor.versionType": {"branch"},
	}
	if withContent {
		query.Set("includeContent", "true")
	}
	var found item
	err := s.getJSON(ctx, s.base(), "/items", query, &found)
	if errors.Is(err, errNotFound) {
		return item{}, board.ErrNotFound
	}
	return found, err
}

func (s *Source) ReadFile(ctx context.Context, branch, path string) ([]byte, string, error) {
	found, err := s.item(ctx, branch, path, true)
	if err != nil {
		return nil, "", err
	}
	return []byte(found.Content), found.ObjectID, nil
}

func (s *Source) WriteFile(ctx context.Context, branch, path string, content []byte, version, message string) error {
	for range pushAttempts {
		head, err := s.Head(ctx, branch)
		if err != nil {
			return err
		}
		current, err := s.item(ctx, branch, path, false)
		if err != nil {
			return err
		}
		if current.ObjectID != version {
			return board.ErrChanged
		}
		pushed, err := s.push(ctx, branch, head, path, content, message)
		if err != nil {
			return err
		}
		if pushed {
			return nil
		}
		moved, err := s.Head(ctx, branch)
		if err != nil {
			return err
		}
		if moved == head {
			return errPushRefused
		}
	}
	return board.ErrChanged
}

func (s *Source) push(ctx context.Context, branch, head, path string, content []byte, message string) (bool, error) {
	payload, err := json.Marshal(map[string]any{
		"refUpdates": []map[string]string{{"name": pushBranchPrefix + branch, "oldObjectId": head}},
		"commits": []map[string]any{{
			"comment": message,
			"changes": []map[string]any{{
				"changeType": "edit",
				"item":       map[string]string{"path": "/" + path},
				"newContent": map[string]string{"content": string(content), "contentType": "rawtext"},
			}},
		}},
	})
	if err != nil {
		return false, err
	}
	res, err := s.send(ctx, http.MethodPost, s.base()+"/pushes", nil, payload)
	if err != nil {
		return false, err
	}
	return res.status < 300, nil
}

func (s *Source) LastEditor(ctx context.Context, branch, path string) (string, error) {
	var commits struct {
		Value []struct {
			Author struct {
				Name string `json:"name"`
			} `json:"author"`
		} `json:"value"`
	}
	err := s.getJSON(ctx, s.base(), "/commits", url.Values{
		"searchCriteria.itemPath":                {"/" + path},
		"searchCriteria.itemVersion.version":     {branch},
		"searchCriteria.itemVersion.versionType": {"branch"},
		"$top":                                   {"1"},
	}, &commits)
	if err != nil || len(commits.Value) == 0 {
		return "", err
	}
	return commits.Value[0].Author.Name, nil
}

func (s *Source) CanWrite(ctx context.Context) (bool, error) {
	var repo struct {
		ID      string `json:"id"`
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	}
	if err := s.getJSON(ctx, s.base(), "", nil, &repo); err != nil {
		return false, readOnlyOn(err)
	}
	var allowed struct {
		Value []bool `json:"value"`
	}
	err := s.getJSON(ctx, s.host+"/"+url.PathEscape(s.repo.Org), "/_apis/permissions/"+gitNamespace+"/"+contributeBit, url.Values{
		"tokens": {"repoV2/" + repo.Project.ID + "/" + repo.ID},
	}, &allowed)
	if err != nil {
		return false, readOnlyOn(err)
	}
	return len(allowed.Value) == 1 && allowed.Value[0], nil
}
