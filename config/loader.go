package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	Config       = viper.New()
	configError  error
	configMu     sync.Mutex
	configLoaded bool
)

// Load resolves and reads the active Nuwa config once per process.
func Load() error {
	configMu.Lock()
	if configLoaded {
		err := configError
		configMu.Unlock()
		return err
	}
	configLoaded = true
	configError = nil
	Config = viper.New()
	configMu.Unlock()

	configPath, err := resolveConfigPath()
	if err != nil {
		setConfigError(err)
		return err
	}

	Config.AddConfigPath(configPath)
	Config.SetConfigName(Environment())
	Config.SetConfigType("yaml")
	if err := Config.ReadInConfig(); err != nil {
		setConfigError(err)
		return err
	}

	if err := Config.UnmarshalKey("redis", &Redis); err != nil {
		setConfigError(err)
		return err
	}
	if err := Config.UnmarshalKey("clickhouse", &ClickHouse); err != nil {
		setConfigError(err)
		return err
	}

	if os.Getenv("NUWA_CONFIG_WATCH") == "true" {
		Config.WatchConfig()
		Config.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("Config file changed: %s\n", e.Name)
		})
	}

	return nil
}

func ConfigError() error {
	configMu.Lock()
	defer configMu.Unlock()
	return configError
}

// SetConfigErrorForTest allows tests in dependent packages to override load-time config errors.
func SetConfigErrorForTest(err error) {
	configMu.Lock()
	defer configMu.Unlock()

	configError = err
	if err == nil {
		configLoaded = false
		Config = viper.New()
		Redis = nil
		ClickHouse = nil
		return
	}
	configLoaded = true
}

// Environment retrieves the configuration environment name, defaulting to "dev".
func Environment() string {
	env := os.Getenv("environment")
	switch env {
	case "prod":
		return "prod"
	case "test":
		return "test"
	default:
		return "dev"
	}
}

func resolveConfigPath() (string, error) {
	if path := os.Getenv("NUWA_CONFIG_DIR"); path != "" {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return "", fmt.Errorf("config path is not a directory: %s", path)
		}
		return path, nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := workDir
	for {
		configPath := filepath.Join(dir, "config")
		info, err := os.Stat(configPath)
		if err == nil && info.IsDir() {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config directory not found from %s", workDir)
}

func setConfigError(err error) {
	configMu.Lock()
	configError = err
	configMu.Unlock()
}
