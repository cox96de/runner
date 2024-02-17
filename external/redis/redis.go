package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	conn *redis.Client
}

func (c *Client) Lock(ctx context.Context, key, value string, expiresIn time.Duration) (bool, error) {
	ok, err := c.conn.SetNX(ctx, key, value, expiresIn).Result()
	if err != nil {
		return false, err
	}
	return ok, err
}

func (c *Client) Unlock(ctx context.Context, key string) (bool, error) {
	result, err := c.conn.Del(ctx, key).Result()
	return result == 1, err
}

func NewClient(conn *redis.Client) *Client {
	return &Client{conn: conn}
}
