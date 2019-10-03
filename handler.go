package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type handler struct {
	router *mux.Router
	tmpl   map[string]*template.Template
}

func NewHandler() (*handler, error) {
	h := &handler{
		router: mux.NewRouter(),
		tmpl:   make(map[string]*template.Template),
	}

	h.router.StrictSlash(true)

	static := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	h.router.PathPrefix("/static/").Handler(static)

	h.router.HandleFunc("/", h.showHome)
	h.router.HandleFunc("/{repository}/", h.showTree)
	h.router.HandleFunc("/{repository}/commits", h.showCommits)
	h.router.HandleFunc("/{repository}/tree/{path:.*}", h.showTree)
	h.router.HandleFunc("/{repository}/blob/{path:.*}", h.showBlob)
	h.router.HandleFunc("/{repository}/raw/{path:.*}", h.sendBlob)

	pages := []string{"home", "commits", "tree", "blob"}
	for _, page := range pages {
		path := fmt.Sprintf("template/%s.html", page)

		t, err := template.ParseFiles("template/layout.html", path)
		if err != nil {
			return nil, err
		}

		h.tmpl[page] = t
	}

	return h, nil
}

func (h *handler) showError(w http.ResponseWriter, r *http.Request, status int, err error) {
	switch status {
	case http.StatusNotFound:
		http.NotFound(w, r)
	case http.StatusInternalServerError:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) showHome(w http.ResponseWriter, r *http.Request) {
	names, err := getRepositoryNames()
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := struct {
		Names []string
	}{
		names,
	}

	err = h.tmpl["home"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *handler) showCommits(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	commits, err := getRepositoryCommits(repository)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := struct {
		Name    string
		Commits []*object.Commit
	}{
		vars["repository"],
		commits,
	}

	err = h.tmpl["commits"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *handler) showTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	path, ok := vars["path"]
	if !ok {
		path = ""
	}

	tree, err := getRepositoryTree(repository, path)
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	objects, err := getTreeObjects(tree)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	params := struct {
		Name    string
		Path    string
		Objects []*treeObject
	}{
		vars["repository"],
		path,
		objects,
	}

	err = h.tmpl["tree"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *handler) showBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	tree, err := getRepositoryTree(repository, "")
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	file, err := tree.File(vars["path"])
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	binary, err := file.IsBinary()
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	var contents string
	if !binary {
		reader, err := file.Blob.Reader()
		if err != nil {
			h.showError(w, r, http.StatusInternalServerError, err)
			return
		}

		contents, err = highlight(file.Name, reader)
		if err != nil {
			h.showError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	params := struct {
		Name     string
		Path     string
		Binary   bool
		Contents template.HTML
	}{
		vars["repository"],
		vars["path"],
		binary,
		template.HTML(contents),
	}

	err = h.tmpl["blob"].ExecuteTemplate(w, "layout", params)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (h *handler) sendBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	tree, err := getRepositoryTree(repository, "")
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	file, err := tree.File(vars["path"])
	if err != nil {
		h.showError(w, r, http.StatusNotFound, nil)
		return
	}

	reader, err := file.Blob.Reader()
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	binary, err := file.IsBinary()
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}

	if binary {
		w.Header().Set("Content-Type", "application/octet-stream")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	_, err = io.Copy(w, reader)
	if err != nil {
		h.showError(w, r, http.StatusInternalServerError, err)
		return
	}
}
