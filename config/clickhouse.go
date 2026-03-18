package config

var (
	ClickHouse *ClickHouseConfig
)

type (
	ClickHouseConfig struct {
		Address  string `mapstructure:"address"`
		Password string `mapstructure:"password"`
		Username string `mapstructure:"username"`
	}
)
