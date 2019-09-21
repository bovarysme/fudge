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

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type fileInfo struct {
	Name   string
	IsFile bool
	Size   string
}

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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	names, err := getRepositoryNames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("template/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := struct {
		Names []string
	}{
		names,
	}
	t.Execute(w, params)
}

func repositoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ok, err := isRepository(vars["repository"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(root, vars["repository"])
	repository, err := git.PlainOpen(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	head, err := repository.Head()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commit, err := repository.CommitObject(head.Hash())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	walker := object.NewTreeWalker(tree, false, nil)

	var files []*fileInfo
	for {
		name, entry, err := walker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		size, err := tree.Size(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		info := &fileInfo{
			Name:   name,
			IsFile: entry.Mode.IsFile(),
			Size:   humanize.Bytes(uint64(size)),
		}

		files = append(files, info)
	}

	t, err := template.ParseFiles("template/repository.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := struct {
		Name  string
		Files []*fileInfo
	}{
		vars["repository"],
		files,
	}
	t.Execute(w, params)
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
