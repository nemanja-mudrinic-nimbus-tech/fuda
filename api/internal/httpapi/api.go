package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"fuda/internal/board"
	"fuda/internal/guide"
)

type api struct {
	log    *slog.Logger
	boards *board.Boards
	logins []Login
}

type boardHost struct {
	name   string
	params []string
}

var boardHosts = []boardHost{
	{"github", []string{"owner", "repo"}},
	{"azure", []string{"org", "project", "repo"}},
	{"local", []string{"folder"}},
}

type boardHandler func(w http.ResponseWriter, r *http.Request, service *board.Service)

func (a api) routes(mux *http.ServeMux) {
	for _, host := range boardHosts {
		prefix := "/api/" + host.name
		for _, param := range host.params {
			prefix += "/{" + param + "}"
		}
		for pattern, handler := range map[string]boardHandler{
			"GET /board":              a.board,
			"GET /tasks/{id}":         a.task,
			"POST /tasks/{id}/move":   a.move,
			"POST /tasks/{id}/assign": a.assign,
			"GET /search":             a.search,
			"GET /archive":            a.archive,
			"GET /docs":               a.doc,
			"GET /files":              a.file,
			"POST /sync":              a.sync,
		} {
			method, suffix, _ := strings.Cut(pattern, " ")
			mux.HandleFunc(method+" "+prefix+suffix, a.onBoard(host, handler))
		}
	}
	mux.HandleFunc("GET /api/boards", a.boardList)
	mux.HandleFunc("GET /api/guide", a.guideList)
	mux.HandleFunc("GET /api/guide/{slug}", a.guidePage)
}

func (a api) onBoard(host boardHost, handler boardHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := make([]string, len(host.params))
		for i, param := range host.params {
			parts[i] = r.PathValue(param)
		}
		ctx, err := a.withToken(w, r, host.name)
		if err != nil {
			a.result(w, nil, err)
			return
		}
		service, err := a.boards.Get(ctx, board.BoardID{Host: host.name, Repo: strings.Join(parts, "/")})
		if err != nil {
			a.result(w, nil, err)
			return
		}
		handler(w, r.WithContext(ctx), service)
	}
}

func (a api) withToken(w http.ResponseWriter, r *http.Request, host string) (context.Context, error) {
	if host == "local" {
		return r.Context(), nil
	}
	login := a.loginFor(host)
	if login == nil {
		return nil, board.ErrNotFound
	}
	token, err := login.Token(w, r)
	if err != nil {
		return nil, board.ErrUnauthorized
	}
	return board.WithToken(r.Context(), token), nil
}

func (a api) loginFor(host string) Login {
	for _, login := range a.logins {
		if login.Host() == host {
			return login
		}
	}
	return nil
}

type boardListing struct {
	Host  string `json:"host"`
	Repo  string `json:"repo"`
	Path  string `json:"path"`
	Title string `json:"title"`
}

type hostStatus struct {
	Host     string `json:"host"`
	LoggedIn bool   `json:"loggedIn"`
	Login    string `json:"login"`
	Error    string `json:"error,omitempty"`
}

type boardListResponse struct {
	Boards []boardListing `json:"boards"`
	Hosts  []hostStatus   `json:"hosts"`
}

func (a api) boardList(w http.ResponseWriter, r *http.Request) {
	response := boardListResponse{Boards: []boardListing{}, Hosts: []hostStatus{}}
	if ids, err := a.boards.List(r.Context(), "local"); err == nil {
		response.Boards = append(response.Boards, listings(ids)...)
	}
	for _, login := range a.logins {
		status := hostStatus{Host: login.Host(), Login: "/auth/" + login.Host() + "/login"}
		ctx, err := a.withToken(w, r, login.Host())
		if err == nil {
			var ids []board.BoardID
			ids, err = a.boards.List(ctx, login.Host())
			response.Boards = append(response.Boards, listings(ids)...)
		}
		switch {
		case err == nil:
			status.LoggedIn = true
		case errors.Is(err, board.ErrUnauthorized):
		default:
			a.log.Warn("could not list Boards", "host", login.Host(), "error", err)
			status.LoggedIn = true
			status.Error = "could not list Boards"
		}
		response.Hosts = append(response.Hosts, status)
	}
	a.json(w, http.StatusOK, response)
}

func listings(ids []board.BoardID) []boardListing {
	listing := make([]boardListing, 0, len(ids))
	for _, id := range ids {
		listing = append(listing, boardListing{Host: id.Host, Repo: id.Repo, Path: id.Path(), Title: path.Base(id.Repo)})
	}
	return listing
}

func (a api) board(w http.ResponseWriter, r *http.Request, service *board.Service) {
	view, ready := service.Board()
	if !ready {
		a.json(w, http.StatusServiceUnavailable, view)
		return
	}
	canWrite, err := service.CanWrite(r.Context())
	if errors.Is(err, board.ErrUnauthorized) {
		a.result(w, nil, err)
		return
	}
	if err != nil {
		a.log.Warn("write access could not be checked; showing a read-only board", "error", err)
	}
	view.ReadOnly = !canWrite
	a.json(w, http.StatusOK, view)
}

func (a api) task(w http.ResponseWriter, r *http.Request, service *board.Service) {
	view, err := service.Task(r.PathValue("id"))
	a.result(w, view, err)
}

func (a api) move(w http.ResponseWriter, r *http.Request, service *board.Service) {
	var req board.MoveRequest
	if !a.decode(w, r, &req) {
		return
	}
	req.TaskID = r.PathValue("id")
	a.saved(w, service.Move(r.Context(), req))
}

func (a api) assign(w http.ResponseWriter, r *http.Request, service *board.Service) {
	var req board.AssignRequest
	if !a.decode(w, r, &req) {
		return
	}
	req.TaskID = r.PathValue("id")
	a.saved(w, service.Assign(r.Context(), req))
}

func (a api) decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if r.Header.Get("Content-Type") != "application/json" {
		a.json(w, http.StatusUnsupportedMediaType, map[string]string{"error": "send application/json"})
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(into); err != nil {
		a.json(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return false
	}
	return true
}

func (a api) saved(w http.ResponseWriter, err error) {
	if err != nil {
		a.result(w, nil, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a api) search(w http.ResponseWriter, r *http.Request, service *board.Service) {
	a.json(w, http.StatusOK, service.Search(r.URL.Query().Get("q")))
}

func (a api) archive(w http.ResponseWriter, _ *http.Request, service *board.Service) {
	a.json(w, http.StatusOK, service.Archive())
}

func (a api) doc(w http.ResponseWriter, r *http.Request, service *board.Service) {
	view, err := service.Doc(r.URL.Query().Get("path"))
	a.result(w, view, err)
}

func (a api) file(w http.ResponseWriter, r *http.Request, service *board.Service) {
	content, contentType, err := service.Asset(r.URL.Query().Get("path"))
	if err != nil {
		a.result(w, nil, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = w.Write(content)
}

func (a api) guideList(w http.ResponseWriter, _ *http.Request) {
	a.json(w, http.StatusOK, guide.List())
}

func (a api) guidePage(w http.ResponseWriter, r *http.Request) {
	page, err := guide.Render(r.PathValue("slug"))
	if errors.Is(err, guide.ErrNotFound) {
		err = board.ErrNotFound
	}
	a.result(w, page, err)
}

func (a api) sync(w http.ResponseWriter, r *http.Request, service *board.Service) {
	if !service.RequestSync(r.Context()) {
		a.json(w, http.StatusTooManyRequests, map[string]string{"error": "a sync ran moments ago; try again shortly"})
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (a api) result(w http.ResponseWriter, v any, err error) {
	var conflict *board.ConflictError
	switch {
	case errors.Is(err, board.ErrNotFound):
		a.json(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, board.ErrUnauthorized):
		a.json(w, http.StatusUnauthorized, map[string]string{"error": "login required"})
	case errors.Is(err, board.ErrForbidden):
		a.json(w, http.StatusForbidden, map[string]string{"error": "no access"})
	case errors.Is(err, board.ErrLocked):
		a.json(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.As(err, &conflict):
		a.json(w, http.StatusConflict, map[string]string{"error": conflict.Message})
	case err != nil:
		a.log.Error("request failed", "error", err)
		a.json(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	default:
		a.json(w, http.StatusOK, v)
	}
}

func (a api) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		a.log.Error("encode response", "error", err)
	}
}
