package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfig("testdata/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Root != "../" {
		t.Errorf("wrong root value: got %v want %v", cfg.Root, "../")
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
