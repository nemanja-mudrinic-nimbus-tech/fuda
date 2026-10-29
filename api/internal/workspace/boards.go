package workspace

import (
	"context"
	"log/slog"
	"path"
	"path/filepath"

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

func newSource(cfg config.Config, folders *Folders, id board.ID) (board.Source, hostLinks, error) {
	switch id.Host {
	case board.HostLocal:
		folder, ok := folders.Path(id)
		if !ok {
			return nil, hostLinks{}, board.ErrNotFound
		}
		return local.New(folder, docsRoot, boardDir, workBranch, cfg.LocalWrites), hostLinks{origin: board.Origin{Host: board.HostLocal, Repo: id.Repo}}, nil
	case board.HostGitHub:
		if !cfg.Has(board.HostGitHub) {
			return nil, hostLinks{}, board.ErrNotFound
		}
		gh := github.New(id.Repo, docsRoot)
		return gh, hostLinks{
			code:   gh.CodeURL(workBranch),
			pr:     gh.PRLink(),
			origin: board.Origin{Host: board.HostGitHub, Repo: id.Repo, URL: "https://github.com/" + id.Repo},
		}, nil
	case board.HostAzure:
		repo, ok := azure.ParseRepo(id.Repo)
		if !ok || !cfg.Has(board.HostAzure) {
			return nil, hostLinks{}, board.ErrNotFound
		}
		az := azure.New(repo, docsRoot)
		return az, hostLinks{
			code:   az.CodeURL(workBranch),
			pr:     az.PRLink(),
			origin: board.Origin{Host: board.HostAzure, Repo: repo.Name, URL: repo.WebURL()},
		}, nil
	}
	return nil, hostLinks{}, board.ErrNotFound
}

func listBoards(cfg config.Config, folders *Folders) func(context.Context, board.Host) ([]board.ID, error) {
	return func(ctx context.Context, host board.Host) ([]board.ID, error) {
		switch {
		case host == board.HostGitHub && cfg.Has(board.HostGitHub):
			return boardIDs(ctx, board.HostGitHub, cfg.GitHubBoards, github.ListBoards)
		case host == board.HostAzure && cfg.Has(board.HostAzure):
			return boardIDs(ctx, board.HostAzure, cfg.AzureBoards, azure.ListBoards)
		case host == board.HostLocal:
			return folders.List(), nil
		}
		return nil, nil
	}
}

func boardIDs(ctx context.Context, host board.Host, listed []string, list func(context.Context, []string) ([]string, error)) ([]board.ID, error) {
	repos, err := list(ctx, listed)
	if err != nil {
		return nil, err
	}
	ids := make([]board.ID, len(repos))
	for i, repo := range repos {
		ids[i] = board.ID{Host: host, Repo: repo}
	}
	return ids, nil
}

func NewRegistry(log *slog.Logger, cfg config.Config, folders *Folders) *board.Registry {
	home := board.ID{}
	if ids := folders.List(); cfg.Has(board.HostLocal) && len(ids) > 0 {
		home = ids[0]
	}
	installs := map[board.Host]board.AppInstall{}
	if cfg.Has(board.HostGitHub) && cfg.GitHubAppSlug != "" {
		installs[board.HostGitHub] = board.AppInstall{
			URL:       "https://github.com/apps/" + cfg.GitHubAppSlug + "/installations/new",
			Installed: github.AppInstalled,
		}
	}
	return board.NewRegistry(log, func(id board.ID) (*board.Service, error) {
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
			WatchMain:  cfg.WatchMain && id.Host != board.HostLocal,
			Cooldown:   cfg.SyncCooldown,
			CacheDir:   filepath.Join(cfg.CacheDir, string(id.Host), filepath.FromSlash(id.Repo)),
			Logger:     log,
			CodeURL:    links.code,
			PRLink:     links.pr,
			Origin:     links.origin,
		}), nil
	}, listBoards(cfg, folders), installs)
}
