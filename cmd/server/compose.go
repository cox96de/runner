package main

import (
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/external/redis"
	"github.com/cox96de/runner/lib"
	goredis "github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func ComposeDB(c *DB) (*db.Client, error) {
	var (
		conn    *gorm.DB
		err     error
		dialect = c.Dialect
		dsn     = c.DSN
	)
	opts := &gorm.Config{}
	if c.TablePrefix != "" {
		opts.NamingStrategy = &schema.NamingStrategy{
			TablePrefix: c.TablePrefix,
		}
	}
	switch db.Dialect(dialect) {
	case db.Mysql:
		conn, err = gorm.Open(mysql.Open(dsn), opts)
	case db.Postgres:
		conn, err = gorm.Open(postgres.Open(dsn), opts)
	case db.SQLite:
		conn, err = gorm.Open(sqlite.Open(dsn), opts)
	default:
		return nil, errors.Errorf("unsupported dialect: %s", dialect)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open database connection")
	}
	return db.NewClient(db.Dialect(dialect), conn), nil
}

func ComposeLocker(l *Locker) (lib.Locker, error) {
	switch l.Backend {
	case "redis":
		return ComposeRedis(l.Redis), nil
	default:
		return nil, errors.Errorf("%s locker is not supported", l.Backend)
	}
}

func ComposeRedis(r *Redis) *redis.Client {
	conn := goredis.NewClient(&goredis.Options{
		Network:            "",
		Addr:               r.Addr,
		Username:           r.Username,
		Password:           r.Password,
		DB:                 r.DB,
		MaxRetries:         r.MaxRetries,
		MinRetryBackoff:    r.MinRetryBackoff,
		MaxRetryBackoff:    r.MaxRetryBackoff,
		DialTimeout:        r.DialTimeout,
		ReadTimeout:        r.ReadTimeout,
		WriteTimeout:       r.WriteTimeout,
		PoolFIFO:           r.PoolFIFO,
		PoolSize:           r.PoolSize,
		MinIdleConns:       r.MinIdleConns,
		MaxConnAge:         r.MaxConnAge,
		PoolTimeout:        r.PoolTimeout,
		IdleTimeout:        r.IdleTimeout,
		IdleCheckFrequency: r.IdleCheckFrequency,
	})
	return redis.NewClient(conn)
}
