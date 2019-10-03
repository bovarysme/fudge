package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfig("../testdata/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Root != "../" {
		t.Errorf("wrong root value: got %v want %v", cfg.Root, "../")
	}

	if !cfg.Debug {
		t.Errorf("wrong debug value: got %v want %v", cfg.Debug, true)
	}
}
