package git

import (
	"testing"

	"gopkg.in/src-d/go-git.v4"
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
			t.Errorf("error when openning %s/%s (strict: %v): got %v want %v",
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
