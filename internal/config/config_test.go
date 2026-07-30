package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndGetString(t *testing.T) {
	// Temporarily override UserCacheDir using HOME for the test environment
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Ensure config is loaded
	Load()

	// Since Load() creates the config file if it doesn't exist,
	// let's verify it gets created.
	cacheDir, _ := os.UserCacheDir()
	if cacheDir == "" {
		cacheDir = tmpDir
	}
	configPath := filepath.Join(cacheDir, Name, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s", configPath)
	}

	// Test GetString with nil viper
	v = nil
	if val := GetString("some_key"); val != "" {
		t.Errorf("expected empty string from nil viper, got %s", val)
	}

	// Load again to initialize viper
	Load()
	if val := GetString("non_existent"); val != "" {
		t.Errorf("expected empty string for non_existent key, got %s", val)
	}
}

func TestConstants(t *testing.T) {
	if Name != "mcp-server-duckduckgo" {
		t.Error("bad name")
	}
	if Platform != "DuckDuckGo" {
		t.Error("bad platform")
	}
}
