package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func isRepository(filename string) (bool, error) {
	path := filepath.Join(root, filename)

	file, err := os.Stat(path)
	if err != nil {
		return false, nil
	}

	if !file.IsDir() {
		return false, nil
	}

	_, err = git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func getRepositoryNames() ([]string, error) {
	var names []string

	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		ok, err := isRepository(file.Name())
		if err != nil {
			return nil, err
		}
		if ok {
			names = append(names, file.Name())
		}
	}

	return names, nil
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

func repositoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ok, err := isRepository(vars["repository"])
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	if !ok {
		errorHandler(w, r, http.StatusNotFound, nil)
		return
	}

	path := filepath.Join(root, vars["repository"])
	repository, err := git.PlainOpen(path)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	head, err := repository.Head()
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	commit, err := repository.CommitObject(head.Hash())
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
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
		Objects []*treeObject
	}{
		vars["repository"],
		objects,
	}

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/{repository}", repositoryHandler)
	router.HandleFunc("/{repository}/commits/{branch}", nil)
	router.HandleFunc("/{repository}/tree/{commit}/{path:.*}", nil)
	router.HandleFunc("/{repository}/blob/{commit}/{path:.*}", nil)
	router.HandleFunc("/{repository}/raw/{commit}/{path:.*}", nil)

	fmt.Println("Starting server on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
