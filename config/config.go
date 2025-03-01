package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Channels []Channel `yaml:"channels"`
}

type Channel struct {
	SlackChannel string   `yaml:"slack_channel"`
	Feeds        []string `yaml:"feeds"`
}

func LoadConfig(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
