package main

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/external/redis"
	"github.com/cox96de/runner/log"
	goredis "github.com/redis/go-redis/v9"
)

// ComposeInternalRedis returns a redis client with internal redis.
func ComposeInternalRedis() (*redis.Client, error) {
	log.Warningf("You are using internal redis, it is not recommended for production.")
	miniRedis := miniredis.NewMiniRedis()
	if err := miniRedis.Start(); err != nil {
		return nil, err
	}
	conn := goredis.NewClient(&goredis.Options{
		Network: "",
		Addr:    miniRedis.Addr(),
	})
	return redis.NewClient(conn)
}

func DetectSQLiteAndMigrate(dbCli *db.Client) error {
	// TODO: prepare DB in docker-entrypoint.sh.
	log.Infof("try to migrate sqlite database")
	if err := dbCli.AutoMigrate(); err != nil {
		return errors.WithMessage(err, "failed to migrate database")
	}
	return nil
}
