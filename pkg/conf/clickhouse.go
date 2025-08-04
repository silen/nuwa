package conf

var (
	ClickHouse *clickHouseConfig
)

type (
	clickHouseConfig struct {
		Address  string `mapstructure:"address"`
		Password string `mapstructure:"password"`
		Username string `mapstructure:"username"`
	}
)
