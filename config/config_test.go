package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfig("testdata/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	want := "fudge.example.org"
	if cfg.Domain != want {
		t.Errorf("wrong domain value: got %v want %v", cfg.Domain, want)
	}

	want = "https://git.example.org"
	if cfg.GitURL != want {
		t.Errorf("wrong git-url value: got %v want %v", cfg.GitURL, want)
	}

	want = "/home/git/"
	if cfg.RepoRoot != want {
		t.Errorf("wrong root value: got %v want %v", cfg.RepoRoot, want)
	}

	if !cfg.Debug {
		t.Errorf("wrong debug value: got %v want %v", cfg.Debug, true)
	}

	descriptions := map[string]string{
		"simple":     "A simple description",
		"multi-line": "A multiline description.\nThis is the second line.\n",
	}
	for name, want := range descriptions {
		got, ok := cfg.Descriptions[name]
		if !ok {
			t.Errorf("expected %v to have a description", name)
		}

		if got != want {
			t.Errorf("wrong description for %v: got %v want %v", name, got, want)
		}
	}
}
