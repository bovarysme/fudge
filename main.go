package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func openRepository(filename string) (*git.Repository, error) {
	path := filepath.Join(root, filename)
	repository, err := git.PlainOpen(path)

	return repository, err
}

func getRepositoryNames() ([]string, error) {
	var names []string

	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		_, err := openRepository(file.Name())
		if err == git.ErrRepositoryNotExists {
			continue
		}
		if err != nil {
			return nil, err
		}

		names = append(names, file.Name())
	}

	return names, nil
}

func getRepositoryTree(repository *git.Repository, path string) (*object.Tree, error) {
	head, err := repository.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repository.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	if path != "" {
		tree, err = tree.Tree(path)
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

type treeObject struct {
	Name   string
	IsFile bool
	Size   string // The object humanized size
}

func getTreeObjects(tree *object.Tree) ([]*treeObject, error) {
	var objects []*treeObject

	walker := object.NewTreeWalker(tree, false, nil)

	for {
		name, entry, err := walker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		size, err := tree.Size(name)
		if err != nil {
			return nil, err
		}

		o := &treeObject{
			Name:   name,
			IsFile: entry.Mode.IsFile(),
			Size:   humanize.Bytes(uint64(size)),
		}

		objects = append(objects, o)
	}

	sort.Slice(objects, func(i, j int) bool {
		// Order the objects by non-files then names
		if !objects[i].IsFile && objects[j].IsFile {
			return true
		}

		if objects[i].IsFile && !objects[j].IsFile {
			return false
		}

		return objects[i].Name < objects[j].Name
	})

	return objects, nil
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int, err error) {
	switch status {
	case http.StatusNotFound:
		http.NotFound(w, r)
	case http.StatusInternalServerError:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	names, err := getRepositoryNames()
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("template/layout.html", "template/home.html")
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	params := struct {
		Names []string
	}{
		names,
	}

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}
}

func getCommits(r *git.Repository) ([]*object.Commit, error) {
	iter, err := r.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit

	// TODO: paginate using NewFilterCommitIter
	err = iter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func commitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	commits, err := getCommits(repository)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("template/layout.html", "template/commits.html")
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	params := struct {
		Name    string
		Commits []*object.Commit
	}{
		vars["repository"],
		commits,
	}

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}
}

func repositoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	path, ok := vars["path"]
	if !ok {
		path = ""
	}

	tree, err := getRepositoryTree(repository, path)
	if err != nil {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}

	objects, err := getTreeObjects(tree)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	t, err := template.ParseFiles("template/layout.html", "template/repository.html")
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
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

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}
}

func blobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	repository, err := openRepository(vars["repository"])
	if err == git.ErrRepositoryNotExists {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	tree, err := getRepositoryTree(repository, "")
	if err != nil {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}

	file, err := tree.File(vars["path"])
	if err != nil {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}

	binary, err := file.IsBinary()
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	var contents string
	if !binary {
		reader, err := file.Blob.Reader()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err)
			return
		}

		contents, err = highlight(file.Name, reader)
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	t, err := template.ParseFiles("template/layout.html", "template/blob.html")
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
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

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}
}

func main() {
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))

	router := mux.NewRouter()
	router.StrictSlash(true)

	// TODO: favicon.ico, robots.txt routes
	router.PathPrefix("/static/").Handler(staticHandler)
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/{repository}/", repositoryHandler)
	router.HandleFunc("/{repository}/commits", commitHandler)
	router.HandleFunc("/{repository}/tree/{path:.*}", repositoryHandler)
	router.HandleFunc("/{repository}/blob/{path:.*}", blobHandler)
	router.HandleFunc("/{repository}/raw/{path:.*}", nil)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	server := &http.Server{
		Addr:         "localhost:8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger,
	}

	logger.Println("Starting server on", server.Addr)
	log.Fatal(server.ListenAndServe())
}
