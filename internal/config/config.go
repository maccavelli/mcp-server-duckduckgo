package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
)

const (
	Name     = "mcp-server-duckduckgo"
	Platform = "DuckDuckGo"

	DefaultTimeout   = 15 * time.Second
	MaxBodyBytes     = 10 * 1024 * 1024
	MaxSnippetLength = 1000
)

const defaultTemplate = `
# mcp-server-duckduckgo Configuration
# Set configurations here
`

var (
	mu sync.RWMutex
	v  *viper.Viper
)

func Load() {
	mu.Lock()
	defer mu.Unlock()

	v = viper.New()

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	configDir := filepath.Join(cacheDir, Name)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return
	}

	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.WriteFile(configPath, []byte(defaultTemplate), 0644)
	}

	v.SetConfigFile(configPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read config: %v\n", err)
	}
}

func GetString(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if v == nil {
		return ""
	}
	return v.GetString(key)
}
