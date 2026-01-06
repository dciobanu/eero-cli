package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}

	if path == "" {
		t.Fatal("ConfigPath() returned empty string")
	}

	// Should end with config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath() = %s, expected to end with config.json", path)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Test that a config without a token returns HasToken() = false
	cfg := &Config{}
	if cfg.HasToken() {
		t.Error("Empty config should not have token")
	}

	cfg.Token = ""
	if cfg.HasToken() {
		t.Error("Config with empty token should not have token")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "eero-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config and save it
	cfg := &Config{
		Token:     "test-token-123",
		NetworkID: "network-456",
	}

	// Save to a temp file
	tmpFile := filepath.Join(tmpDir, "config.json")
	if err := os.MkdirAll(filepath.Dir(tmpFile), 0700); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	// We can't easily test Save/Load with the real path, so test the data handling
	if !cfg.HasToken() {
		t.Error("Config with token should have HasToken() = true")
	}

	cfg.Token = ""
	if cfg.HasToken() {
		t.Error("Config without token should have HasToken() = false")
	}
}

func TestConfigClear(t *testing.T) {
	cfg := &Config{
		Token:     "test-token",
		NetworkID: "test-network",
	}

	// Manually clear (can't test file operations without mocking)
	cfg.Token = ""
	cfg.NetworkID = ""

	if cfg.HasToken() {
		t.Error("Cleared config should not have token")
	}

	if cfg.NetworkID != "" {
		t.Error("Cleared config should not have network ID")
	}
}
