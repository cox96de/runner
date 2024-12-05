package main

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/log"
	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
)

// ComposeInternalRedis returns a redis client with internal redis.
func ComposeInternalRedis() (*goredis.Client, error) {
	log.Warningf("You are using internal redis, it is not recommended for production.")
	miniRedis := miniredis.NewMiniRedis()
	if err := miniRedis.Start(); err != nil {
		return nil, err
	}
	conn := goredis.NewClient(&goredis.Options{
		Network: "",
		Addr:    miniRedis.Addr(),
	})
	if err := redisotel.InstrumentTracing(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument tracing")
	}
	if err := redisotel.InstrumentMetrics(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument metrics")
	}
	return conn, nil
}
