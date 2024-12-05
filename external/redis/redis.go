package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func (c *Client) Lock(ctx context.Context, key, value string, expiresIn time.Duration) (bool, error) {
	ok, err := c.Client.SetNX(ctx, key, value, expiresIn).Result()
	if err != nil {
		return false, err
	}
	return ok, err
}

func (c *Client) Unlock(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Del(ctx, key).Result()
	return result == 1, err
}

func NewClient(conn *redis.Client) *Client {
	return &Client{Client: conn}
}
