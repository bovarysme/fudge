package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Domain       string            `yaml:"domain"`
	GitURL       string            `yaml:"git-url"`
	RepoRoot     string            `yaml:"repo-root"`
	Debug        bool              `yaml:"debug"`
	Descriptions map[string]string `yaml:"descriptions"`
}

func NewConfig(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
