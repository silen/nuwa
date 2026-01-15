package conf

var (
	Redis *redisConfig
)

type (
	redisConfig struct {
		Host         string `mapstructure:"host"`
		Password     string `mapstructure:"password"`
		PoolSize     int    `mapstructure:"pool_size"`
		DB           int    `mapstructure:"db"`
		MinIdleConn  int    `mapstructure:"min_idle_conn"`      // 最小空闲连接数
		MaxIdleConn  int    `mapstructure:"max_idle_conn"`      // 最大空闲连接数
		MaxConnAge   int    `mapstructure:"max_conn_age"`       // 连接最大存活时间（秒）
		ReadTimeout  int    `mapstructure:"read_timeout"`       // 读取超时（秒）
		WriteTimeout int    `mapstructure:"write_timeout"`      // 写入超时（秒）
		DialTimeout  int    `mapstructure:"dial_timeout"`       // 拨号超时（秒）
		IdleTimeout  int    `mapstructure:"idle_timeout"`       // 空闲超时（秒）
	}
)
