package app

import (
	"context"
	"log/slog"
	"path"
	"path/filepath"
	"strings"

	"fuda/internal/board"
	"fuda/internal/config"
	"fuda/internal/source/azure"
	"fuda/internal/source/github"
	"fuda/internal/source/local"
)

const (
	docsRoot   = "docs"
	boardDir   = "docs/board"
	workBranch = "develop"
	prodBranch = "main"
)

type hostLinks struct {
	code   func(string) string
	pr     string
	origin board.Origin
}

func newSource(cfg config.Config, folders *Folders, id board.BoardID) (board.Source, hostLinks, error) {
	switch id.Host {
	case "local":
		folder, ok := folders.Path(id)
		if !ok {
			return nil, hostLinks{}, board.ErrNotFound
		}
		return local.New(folder, docsRoot, boardDir, workBranch), hostLinks{origin: board.Origin{Host: "local", Repo: id.Repo}}, nil
	case "github":
		if !cfg.Has(config.SourceGitHub) {
			return nil, hostLinks{}, board.ErrNotFound
		}
		gh := github.New(id.Repo, docsRoot)
		return gh, hostLinks{
			code:   gh.CodeURL(workBranch),
			pr:     gh.PRLink(),
			origin: board.Origin{Host: "github", Repo: id.Repo, URL: "https://github.com/" + id.Repo},
		}, nil
	case "azure":
		org, project, name, ok := splitAzure(id.Repo)
		if !ok || !cfg.Has(config.SourceAzure) {
			return nil, hostLinks{}, board.ErrNotFound
		}
		repo := azure.Repo{Org: org, Project: project, Name: name}
		az := azure.New(repo, docsRoot)
		return az, hostLinks{
			code:   az.CodeURL(workBranch),
			pr:     az.PRLink(),
			origin: board.Origin{Host: "azure", Repo: name, URL: repo.WebURL()},
		}, nil
	}
	return nil, hostLinks{}, board.ErrNotFound
}

func splitAzure(repo string) (org, project, name string, ok bool) {
	parts := strings.Split(repo, "/")
	if len(parts) != 3 {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

func listBoards(cfg config.Config, folders *Folders) func(context.Context, string) ([]board.BoardID, error) {
	return func(ctx context.Context, host string) ([]board.BoardID, error) {
		switch {
		case host == "github" && cfg.Has(config.SourceGitHub):
			return boardIDs(ctx, "github", github.ListBoards)
		case host == "azure" && cfg.Has(config.SourceAzure):
			return boardIDs(ctx, "azure", azure.ListBoards)
		case host == "local":
			return folders.List(), nil
		}
		return nil, nil
	}
}

func boardIDs(ctx context.Context, host string, list func(context.Context) ([]string, error)) ([]board.BoardID, error) {
	repos, err := list(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]board.BoardID, len(repos))
	for i, repo := range repos {
		ids[i] = board.BoardID{Host: host, Repo: repo}
	}
	return ids, nil
}

func NewBoards(log *slog.Logger, cfg config.Config, folders *Folders) *board.Boards {
	home := board.BoardID{}
	if ids := folders.List(); cfg.Has(config.SourceLocal) && len(ids) > 0 {
		home = ids[0]
	}
	return board.NewBoards(log, func(id board.BoardID) (*board.Service, error) {
		source, links, err := newSource(cfg, folders, id)
		if err != nil {
			return nil, err
		}
		links.origin.Path = id.Path()
		title := path.Base(id.Repo)
		if id == home && cfg.Title != "" {
			title = cfg.Title
		}
		return board.NewService(source, board.Options{
			Title:      title,
			DocsRoot:   docsRoot,
			BoardDir:   boardDir,
			WorkBranch: workBranch,
			ProdBranch: prodBranch,
			WatchMain:  cfg.WatchMain && id.Host != "local",
			Cooldown:   cfg.SyncCooldown,
			CacheDir:   filepath.Join(cfg.CacheDir, id.Host, filepath.FromSlash(id.Repo)),
			Logger:     log,
			CodeURL:    links.code,
			PRLink:     links.pr,
			Origin:     links.origin,
		}), nil
	}, listBoards(cfg, folders))
}
