// @Time : 2020/12/16 4:43 PM
// @Author : silen
// @File : conf
// @Software: vscode
// @Desc: to do somewhat..

package conf

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	Config = viper.New()
)

// getEnvironmentName retrieves the configuration environment name, defaulting to "dev".
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

// init ...
func init() {

	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configPath := filepath.Join(workDir, "config")

	Config.AddConfigPath(configPath)
	Config.SetConfigName(Environment())
	Config.SetConfigType("yaml")
	if err := Config.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	Config.UnmarshalKey("redis", &Redis)

	Config.WatchConfig()
	Config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
	})
}
