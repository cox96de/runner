package main

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/app/server"
	"github.com/cox96de/runner/app/server/eventhook"
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/external/redis"
	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func ComposeDB(c *DB) (*gorm.DB, error) {
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
	if err = conn.Use(otelgorm.NewPlugin()); err != nil {
		return nil, errors.WithMessage(err, "failed to use otelgorm plugin")
	}
	return conn, nil
}

func ComposeLocker(l *Locker) (server.Locker, error) {
	switch l.Backend {
	case "redis":
		r, err := ComposeRedis(l.Redis)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to compose redis")
		}
		return redis.NewClient(r), nil
	default:
		return nil, errors.Errorf("%s locker is not supported", l.Backend)
	}
}

func ComposeLogPersistentStorage(l *LogArchive) (server.LogPersistentStorage, error) {
	var oss server.LogPersistentStorage
	switch l.Backend {
	case "fs":
		oss = logstorage.NewFilesystemOSS(l.BaseDir)
	default:
		return nil, errors.Errorf("%s log archive is not supported", l.Backend)
	}
	return oss, nil
}

func ComposeRedis(r *Redis) (*goredis.Client, error) {
	if r.Addr == "internal" {
		return ComposeInternalRedis()
	}
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
	if err := redisotel.InstrumentTracing(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument tracing")
	}
	if err := redisotel.InstrumentMetrics(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument metrics")
	}
	return conn, nil
}

func ComposeCloudEventsClient(c *Event) (eventhook.Sender, error) {
	switch {
	case c != nil && len(c.HTTPEndPoint) > 0:
		proto, err := cloudevents.NewHTTP(cloudevents.WithTarget(c.HTTPEndPoint))
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to create http client sender")
		}
		http, err := cloudevents.NewClient(proto, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create http client event producer")
		}
		return http, nil
	default:
		return eventhook.NewNopSender(), nil
	}
}
