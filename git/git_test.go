package git

import (
	"testing"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestOpenRepository(t *testing.T) {
	tests := []struct {
		root    string
		dirname string
		strict  bool
		err     error
	}{
		{"testdata/regular_files", "test", true, git.ErrRepositoryNotExists},
		{"testdata/regular_files", "test", false, git.ErrRepositoryNotExists},
		{"testdata/regular_file_with_repo", "test", true, git.ErrRepositoryNotExists},
		{"testdata/regular_file_with_repo", "test.git", true, nil},
		{"testdata/regular_file_with_repo", "test", false, nil},
	}

	for _, test := range tests {
		_, err := OpenRepository(test.root, test.dirname, test.strict)
		if err != test.err {
			t.Errorf("wrong error when openning %s/%s (strict: %v): got %v want %v",
				test.root, test.dirname, test.strict, err, test.err)
		}
	}
}

func TestGetRepositoryNames(t *testing.T) {
	want := []string{"normal", "suffix"}

	got, err := GetRepositoryNames("testdata/repositories")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("wrong slice length: got %d want %d", len(got), len(want))
	}

	for i, name := range got {
		if name != want[i] {
			t.Errorf("wrong repository name: got %s want %s", name, want[i])
		}
	}
}

func TestGetRepositoryCommits(t *testing.T) {
	r, err := OpenRepository("testdata/repository", "python", true)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetRepositoryCommits(r)
	if err != nil {
		t.Fatal(err)
	}

	want := []struct {
		name    string
		when    string
		message string
	}{
		{"Jane Doe", "Oct 24, 2019", "Edit README.md"},
		{"Jane Doe", "Oct 24, 2019", "Add tests"},
		{"Jane Doe", "Oct 24, 2019", "Initial commit"},
	}

	for i, commit := range got {
		if commit.Author.Name != want[i].name {
			t.Errorf("wrong commit author name: got %s want %s",
				commit.Author.Name, want[i].name)
		}

		when := commit.Author.When.Format("Jan 2, 2006")
		if when != want[i].when {
			t.Errorf("wrong commit author when: got %s want %s", when, want[i].when)
		}

		if commit.Message != want[i].message {
			t.Errorf("wrong commit message: got %s want %s",
				commit.Message, want[i].message)
		}
	}
}

func TestGetRepositoryLastCommit(t *testing.T) {
	r, err := OpenRepository("testdata/repository", "python", true)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetRepositoryLastCommit(r)
	if err != nil {
		t.Fatal(err)
	}

	want := struct {
		name    string
		when    string
		message string
	}{
		"Jane Doe",
		"Oct 24, 2019",
		"Edit README.md",
	}

	if got.Author.Name != want.name {
		t.Errorf("wrong commit author name: got %s want %s", got.Author.Name, want.name)
	}

	when := got.Author.When.Format("Jan 2, 2006")
	if when != want.when {
		t.Errorf("wrong commit author when: got %s want %s", when, want.when)
	}

	if got.Message != want.message {
		t.Errorf("wrong commit message: got %s want %s", got.Message, want.message)
	}
}

func TestGetRepositoryTree(t *testing.T) {
	r, err := OpenRepository("testdata/repository", "python", true)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
		err  error
	}{
		{"nonexistent", object.ErrDirectoryNotFound},
		{"/", object.ErrDirectoryNotFound},
		{"src/", object.ErrDirectoryNotFound},
		{"", nil},
		{"src", nil},
		{"src/helpers", nil},
	}

	for _, test := range tests {
		_, err = GetRepositoryTree(r, test.path)
		if err != test.err {
			t.Errorf("wrong error when getting tree %s: got %v want %v",
				test.path, err, test.err)
		}
	}
}
