package git

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/dustin/go-humanize"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type TreeObject struct {
	Name   string
	IsFile bool
	Size   string // The object humanized size
}

type TreeBlob struct {
	Name     string
	IsBinary bool
	Reader   io.ReadCloser
}

func OpenRepository(root, filename string) (*git.Repository, error) {
	path := filepath.Join(root, filename)
	repository, err := git.PlainOpen(path)

	return repository, err
}

func GetRepositoryNames(root string) ([]string, error) {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var names []string

	for _, file := range files {
		_, err := OpenRepository(root, file.Name())
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

func GetRepositoryCommits(r *git.Repository) ([]*object.Commit, error) {
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

func GetRepositoryTree(repository *git.Repository, path string) (*object.Tree, error) {
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

func GetRepositoryBlob(repository *git.Repository, path string) (*TreeBlob, error) {
	dir := filepath.Dir(path)
	if dir == "." {
		dir = ""
	}

	tree, err := GetRepositoryTree(repository, dir)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(path)
	file, err := tree.File(filename)
	if err != nil {
		return nil, err
	}

	isBinary, err := file.IsBinary()
	if err != nil {
		return nil, err
	}

	reader, err := file.Blob.Reader()
	if err != nil {
		return nil, err
	}

	blob := &TreeBlob{
		Name:     file.Name,
		IsBinary: isBinary,
		Reader:   reader,
	}

	return blob, nil
}

func GetTreeObjects(tree *object.Tree) ([]*TreeObject, error) {
	var objects []*TreeObject

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

		o := &TreeObject{
			Name:   name,
			IsFile: entry.Mode.IsFile(),
			Size:   humanize.Bytes(uint64(size)),
		}

		objects = append(objects, o)
	}

	sort.Slice(objects, func(i, j int) bool {
		// Order the objects by non-file status then name
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
