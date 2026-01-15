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
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/silen/nuwa/pkg/logs"
	"github.com/spf13/viper"
)

var (
	Config = viper.New()
	once   sync.Once
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

// InitConfig 初始化配置
func InitConfig() {
	once.Do(func() {
		workDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		configPath := filepath.Join(workDir, "config")

		Config.AddConfigPath(configPath)
		Config.SetConfigName(Environment())
		Config.SetConfigType("yaml")

		if err := Config.ReadInConfig(); err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}

		// 初始化各个配置项
		if err := loadConfigs(); err != nil {
			log.Fatalf("Failed to load configurations: %v", err)
		}

		// 启动配置热更新监听
		watchConfig()
	})
}

// loadConfigs 加载所有配置项
func loadConfigs() error {
	// Redis 配置
	if err := Config.UnmarshalKey("redis", &Redis); err != nil {
		return fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	// 验证配置
	if err := validateConfigs(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// validateConfigs 验证配置的有效性
func validateConfigs() error {
	// 验证 Redis 配置
	if Redis != nil {
		if Redis.Host == "" {
			logs.Warn("Redis host is not configured")
		}
		if Redis.PoolSize <= 0 {
			Redis.PoolSize = 10 // 设置默认值
		}
		if Redis.MinIdleConn <= 0 {
			Redis.MinIdleConn = 5 // 设置默认值
		}
		if Redis.MaxIdleConn <= 0 {
			Redis.MaxIdleConn = 20 // 设置默认值
		}
		if Redis.MaxConnAge <= 0 {
			Redis.MaxConnAge = 3600 // 设置默认值（1小时）
		}
		if Redis.DialTimeout <= 0 {
			Redis.DialTimeout = 5 // 设置默认值（5秒）
		}
		if Redis.ReadTimeout <= 0 {
			Redis.ReadTimeout = 3 // 设置默认值（3秒）
		}
		if Redis.WriteTimeout <= 0 {
			Redis.WriteTimeout = 3 // 设置默认值（3秒）
		}
		if Redis.IdleTimeout <= 0 {
			Redis.IdleTimeout = 300 // 设置默认值（5分钟）
		}
	}

	return nil
}

// watchConfig 监听配置变更
func watchConfig() {
	Config.WatchConfig()
	Config.OnConfigChange(func(e fsnotify.Event) {
		logs.Info("Config file changed: ", e.Name)

		// 重新加载配置
		if err := loadConfigs(); err != nil {
			logs.Error("Failed to reload configs after change: ", err)
		} else {
			logs.Info("Configs reloaded successfully after change")
		}
	})
}

// GetConfig 获取指定键的配置值
func GetConfig(key string) interface{} {
	InitConfig() // 确保配置已初始化
	return Config.Get(key)
}

// GetString 获取字符串类型的配置值
func GetString(key string) string {
	InitConfig()
	return Config.GetString(key)
}

// GetInt 获取整数类型的配置值
func GetInt(key string) int {
	InitConfig()
	return Config.GetInt(key)
}

// GetBool 获取布尔类型的配置值
func GetBool(key string) bool {
	InitConfig()
	return Config.GetBool(key)
}

// GetStringMapString 获取嵌套的字符串映射配置
func GetStringMapString(key string) map[string]string {
	InitConfig()
	return Config.GetStringMapString(key)
}

// IsSet 检查配置键是否存在
func IsSet(key string) bool {
	InitConfig()
	return Config.IsSet(key)
}

// init ...
func init() {
	InitConfig()
}
