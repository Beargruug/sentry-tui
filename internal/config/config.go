// Package config handles loading, saving, and validating application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = "sentry-tui.yaml"
	appDirName     = "sentry-tui"
)

// Config holds all application configuration.
type Config struct {
	AuthToken      string `yaml:"auth_token"`
	Organization   string `yaml:"organization"`
	DefaultProject string `yaml:"default_project,omitempty"`
	BaseURL        string `yaml:"base_url,omitempty"`
	RefreshSeconds int    `yaml:"refresh_seconds,omitempty"`
	Theme          string `yaml:"theme,omitempty"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:        "https://sentry.io/api/0",
		RefreshSeconds: 30,
		Theme:          "dark",
	}
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", appDirName)
	return filepath.Join(dir, configFileName), nil
}

// Load reads the config file and merges with environment variables.
// Environment variables take precedence over file values.
func Load() (Config, error) {
	cfg := DefaultConfig()

	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file — try env vars only
			mergeEnv(&cfg)
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Fill defaults for optional fields if missing
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://sentry.io/api/0"
	}
	if cfg.RefreshSeconds == 0 {
		cfg.RefreshSeconds = 30
	}
	if cfg.Theme == "" {
		cfg.Theme = "dark"
	}

	mergeEnv(&cfg)
	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Exists checks if the config file exists.
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// IsValid checks if minimum required fields are set.
func (c Config) IsValid() bool {
	return c.AuthToken != "" && c.Organization != ""
}

// mergeEnv overrides config values with environment variables if set.
func mergeEnv(cfg *Config) {
	if v := os.Getenv("SENTRY_AUTH_TOKEN"); v != "" {
		cfg.AuthToken = v
	}
	if v := os.Getenv("SENTRY_ORG"); v != "" {
		cfg.Organization = v
	}
	if v := os.Getenv("SENTRY_PROJECT"); v != "" {
		cfg.DefaultProject = v
	}
	if v := os.Getenv("SENTRY_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
}
