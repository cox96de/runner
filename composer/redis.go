package composer

import (
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	Addr            string        `mapstructure:"addr" yaml:"addr"`
	Username        string        `mapstructure:"username" yaml:"username"`
	Password        string        `mapstructure:"password" yaml:"password"`
	DB              int           `mapstructure:"db" yaml:"db"`
	MaxRetries      int           `mapstructure:"max_retries" yaml:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff" yaml:"max_retry_backoff"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	PoolFIFO        bool          `mapstructure:"pool_fifo" yaml:"pool_fifo"`
	PoolSize        int           `mapstructure:"pool_size" yaml:"pool_size"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout" yaml:"pool_timeout"`
	MinIdleConns    int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	MaxActiveConns  int           `mapstructure:"max_active_conns" yaml:"max_active_conns"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

func ComposeRedis(r *Redis) *goredis.Client {
	conn := goredis.NewClient(&goredis.Options{
		Addr:                  r.Addr,
		Username:              r.Username,
		Password:              r.Password,
		DB:                    r.DB,
		MaxRetries:            r.MaxRetries,
		MinRetryBackoff:       r.MinRetryBackoff,
		MaxRetryBackoff:       r.MaxRetryBackoff,
		DialTimeout:           r.DialTimeout,
		ReadTimeout:           r.ReadTimeout,
		WriteTimeout:          r.WriteTimeout,
		ContextTimeoutEnabled: true,
		PoolFIFO:              r.PoolFIFO,
		PoolSize:              r.PoolSize,
		PoolTimeout:           r.PoolTimeout,
		MinIdleConns:          r.MinIdleConns,
		MaxIdleConns:          r.MaxIdleConns,
		MaxActiveConns:        r.MaxActiveConns,
		ConnMaxIdleTime:       r.ConnMaxIdleTime,
		ConnMaxLifetime:       r.ConnMaxLifetime,
	})
	return conn
}
