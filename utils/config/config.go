package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	APIKey     string `yaml:"apiKey"`
	Timeout    int    `yaml:"timeout"`
	MaxRetries int    `yaml:"maxRetries"`
	Model      string `yaml:"model"`
}

func ReadConfig(fileName string) (*Config, error) {

	// Read the file into a byte slice
	buf, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML data into a Config object
	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
