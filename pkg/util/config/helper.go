package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Gateway struct {
		ExcludedPaths []string          `yaml:"excludedPaths"`
		SgPortalURL   string            `yaml:"sgPortalURL"`
		KeyFile       string            `yaml:"keyFile"`
		Services      map[string]string `yaml:"services"`
	} `yaml:"gateway"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadEncryptionKey(path string) ([]byte, error) {
	return os.ReadFile(path)
}
