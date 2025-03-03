package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test valid config
	t.Run("valid config file", func(t *testing.T) {
		// Create a temporary config file
		content := `channels:
  - slack_channel: test-channel
    feeds:
      - https://example.com/feed.xml
      - https://example.com/feed2.xml`

		tmpfile, err := os.CreateTemp("", "config*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		// Test loading the config
		cfg, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Errorf("LoadConfig() error = %v", err)
			return
		}

		// Verify the config contents
		if len(cfg.Channels) != 1 {
			t.Errorf("Expected 1 channel, got %d", len(cfg.Channels))
		}
		if cfg.Channels[0].SlackChannel != "test-channel" {
			t.Errorf("Expected channel name 'test-channel', got '%s'", cfg.Channels[0].SlackChannel)
		}
		if len(cfg.Channels[0].Feeds) != 2 {
			t.Errorf("Expected 2 feeds, got %d", len(cfg.Channels[0].Feeds))
		}
	})

	// Test missing file
	t.Run("missing config file", func(t *testing.T) {
		_, err := LoadConfig("nonexistent.yaml")
		if err == nil {
			t.Error("Expected error for missing file, got nil")
		}
	})

	// Test invalid YAML
	t.Run("invalid yaml", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "config*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte("invalid: yaml: content:")); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = LoadConfig(tmpfile.Name())
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("empty channel name", func(t *testing.T) {
		content := `channels:
  - slack_channel: ""
    feeds:
      - https://example.com/feed.xml`

		tmpfile, err := os.CreateTemp("", "config*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = LoadConfig(tmpfile.Name())
		if err == nil {
			t.Error("Expected error for empty channel name, got nil")
		}
	})
}
