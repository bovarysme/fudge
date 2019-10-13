package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"fudge/config"
	"fudge/git"
	"fudge/util"

	"github.com/gorilla/mux"
	gogit "gopkg.in/src-d/go-git.v4"
)

type Handler struct {
	Router *mux.Router

	config *config.Config
	tmpl   map[string]*template.Template
}

func NewHandler(cfg *config.Config) (*Handler, error) {
	h := &Handler{
		Router: mux.NewRouter(),

		config: cfg,
		tmpl:   make(map[string]*template.Template),
	}

	h.Router.StrictSlash(true)

	static := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	h.Router.PathPrefix("/static/").Handler(static)

	h.Router.HandleFunc("/", h.showHome)
	h.Router.HandleFunc("/{repository}/", h.showTree)
	h.Router.HandleFunc("/{repository}/commits", h.showCommits)
	h.Router.HandleFunc("/{repository}/tree/{path:.*}", h.showTree)
	h.Router.HandleFunc("/{repository}/blob/{path:.*}", h.showBlob)
	h.Router.HandleFunc("/{repository}/raw/{path:.*}", h.sendBlob)

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

	err = h.tmpl["home"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) showCommits(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := git.OpenRepository(h.config.RepoRoot, vars["repository"])
	if err == gogit.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	commits, err := git.GetRepositoryCommits(repository)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := h.getParams(r)

	params["Commits"] = commits

	err = h.tmpl["commits"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) showTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := git.OpenRepository(h.config.RepoRoot, vars["repository"])
	if err == gogit.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

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

	err = h.tmpl["tree"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) showBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := git.OpenRepository(h.config.RepoRoot, vars["repository"])
	if err == gogit.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	blob, err := git.GetRepositoryBlob(repository, vars["path"])
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	var contents string

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

	err = h.tmpl["blob"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) sendBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := git.OpenRepository(h.config.RepoRoot, vars["repository"])
	if err == gogit.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

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
