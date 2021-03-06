package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"bovarys.me/fudge/config"
	"bovarys.me/fudge/git"
	"bovarys.me/fudge/logger"
	"bovarys.me/fudge/util"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	gogit "gopkg.in/src-d/go-git.v4"
)

type Handler struct {
	Router http.Handler

	config *config.Config
	tmpl   map[string]*template.Template
}

func NewHandler(cfg *config.Config) (*Handler, error) {
	h := &Handler{
		config: cfg,
		tmpl:   make(map[string]*template.Template),
	}

	router := mux.NewRouter()
	router.StrictSlash(true)

	static := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	router.PathPrefix("/static/").Handler(static)

	router.HandleFunc("/", h.showHome)
	router.HandleFunc("/{repository}/", h.showTree)
	router.HandleFunc("/{repository}/commits", h.showCommits)
	router.HandleFunc("/{repository}/tree/{path:.*}", h.showTree)
	router.HandleFunc("/{repository}/blob/{path:.*}", h.showBlob)
	router.HandleFunc("/{repository}/raw/{path:.*}", h.sendBlob)

	h.Router = router

	err := h.setLoggers()
	if err != nil {
		return nil, err
	}

	pages := []string{"home", "commits", "tree", "blob", "404", "500"}
	for _, page := range pages {
		path := fmt.Sprintf("template/%s.html", page)

		t, err := template.ParseFiles(
			"template/_layout.html", "template/_utils.html", path)
		if err != nil {
			return nil, err
		}

		h.tmpl[page] = t
	}

	return h, nil
}

func (h *Handler) setLoggers() error {
	loggerConfig, ok := h.config.Loggers["router"]
	if !ok {
		return nil
	}

	if !loggerConfig.Enable {
		return nil
	}

	writer, err := logger.Writer(loggerConfig)
	if err != nil {
		return err
	}

	h.Router = handlers.CombinedLoggingHandler(writer, h.Router)

	return nil
}

func (h *Handler) openRepository(w http.ResponseWriter, r *http.Request) (*gogit.Repository, error) {
	vars := mux.Vars(r)

	repository, err := git.OpenRepository(h.config.RepoRoot, vars["repository"], false)
	if err == gogit.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return nil, err
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return nil, err
	}

	return repository, nil
}

func (h *Handler) getParams(r *http.Request) map[string]interface{} {
	vars := mux.Vars(r)

	repository := vars["repository"]
	path := vars["path"]

	params := make(map[string]interface{})

	params["Domain"] = h.config.Domain
	params["GitURL"] = h.config.GitURL
	params["RepoName"] = repository
	params["Path"] = path

	if repository != "" {
		params["Breadcrumbs"] = util.Breadcrumbs(repository, path)
	}

	return params
}

func (h *Handler) showError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.WriteHeader(status)

	switch status {
	case http.StatusNotFound:
		h.tmpl["404"].ExecuteTemplate(w, "layout", nil)
	case http.StatusInternalServerError:
		params := h.getParams(r)

		params["Debug"] = h.config.Debug
		params["Error"] = err.Error()

		h.tmpl["500"].ExecuteTemplate(w, "layout", params)
	}
}

func (h *Handler) showHome(w http.ResponseWriter, r *http.Request) {
	names, err := git.GetRepositoryNames(h.config.RepoRoot)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := h.getParams(r)

	params["Names"] = names
	params["Descriptions"] = h.config.Descriptions

	h.tmpl["home"].ExecuteTemplate(w, "layout", params)
}

func (h *Handler) showCommits(w http.ResponseWriter, r *http.Request) {
	repository, err := h.openRepository(w, r)
	if err != nil {
		return
	}

	commits, err := git.GetRepositoryCommits(repository)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := h.getParams(r)

	params["Commits"] = commits

	h.tmpl["commits"].ExecuteTemplate(w, "layout", params)
}

func (h *Handler) showTree(w http.ResponseWriter, r *http.Request) {
	repository, err := h.openRepository(w, r)
	if err != nil {
		return
	}

	vars := mux.Vars(r)

	tree, err := git.GetRepositoryTree(repository, vars["path"])
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	objects, err := git.GetTreeObjects(tree)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	commit, err := git.GetRepositoryLastCommit(repository)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := h.getParams(r)

	params["LastCommit"] = commit
	params["Objects"] = objects

	h.tmpl["tree"].ExecuteTemplate(w, "layout", params)
}

func (h *Handler) showBlob(w http.ResponseWriter, r *http.Request) {
	repository, err := h.openRepository(w, r)
	if err != nil {
		return
	}

	vars := mux.Vars(r)

	blob, err := git.GetRepositoryBlob(repository, vars["path"])
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	contents := ""
	if !blob.IsBinary {
		contents, err = util.Highlight(blob.Name, blob.Reader)
		if err != nil {
			h.showError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	commit, err := git.GetRepositoryLastCommit(repository)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := h.getParams(r)

	params["LastCommit"] = commit
	params["Blob"] = blob
	params["Contents"] = template.HTML(contents)

	h.tmpl["blob"].ExecuteTemplate(w, "layout", params)
}

func (h *Handler) sendBlob(w http.ResponseWriter, r *http.Request) {
	repository, err := h.openRepository(w, r)
	if err != nil {
		return
	}

	vars := mux.Vars(r)

	blob, err := git.GetRepositoryBlob(repository, vars["path"])
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	value := "text/plain; charset=utf-8"
	if blob.IsBinary {
		value = "application/octet-stream"
	}

	w.Header().Set("Content-Type", value)

	_, err = io.Copy(w, blob.Reader)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}
