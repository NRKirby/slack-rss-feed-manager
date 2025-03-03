package config

import (
	"errors"
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
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	// Validate config
	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	if len(cfg.Channels) == 0 {
		return errors.New("no channels configured")
	}

	for _, ch := range cfg.Channels {
		if ch.SlackChannel == "" {
			return errors.New("slack channel name cannot be empty")
		}
		if len(ch.Feeds) == 0 {
			return errors.New("no feeds configured for channel " + ch.SlackChannel)
		}
		for _, feed := range ch.Feeds {
			if feed == "" {
				return errors.New("feed URL cannot be empty in channel " + ch.SlackChannel)
			}
		}
	}
	return nil
}
