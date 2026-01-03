// Package config handles configuration file management following XDG conventions
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const (
	appName    = "eero-cli"
	configFile = "config.json"
)

type Config struct {
	Token     string `json:"token"`
	NetworkID string `json:"network_id"`
}

// ConfigPath returns the path to the config file following platform conventions
func ConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, "Library", "Application Support", appName)
	case "linux":
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, appName)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(home, ".config", appName)
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, "."+appName)
	}

	return filepath.Join(configDir, configFile), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// HasToken returns true if a token is configured
func (c *Config) HasToken() bool {
	return c.Token != ""
}

// Clear removes the stored token and network ID
func (c *Config) Clear() error {
	c.Token = ""
	c.NetworkID = ""
	return c.Save()
}
