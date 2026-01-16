package conf

var (
	Redis *redisConfig
)

type (
	redisConfig struct {
		Host     string `mapstructure:"host"`
		Password string `mapstructure:"password"`
		PoolSize int    `mapstructure:"pool_size"`
		DB       int    `mapstructure:"db"`
	}
)
