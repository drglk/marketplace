package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const pkg = "redis/"

type Client struct {
	redisClient *redis.Client
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	res, err := c.redisClient.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}
	return res, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.redisClient.Set(ctx, key, value, expiration).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	err := c.redisClient.Del(ctx, keys...).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	op := pkg + "New"

	client := &Client{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}

	if err := client.redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: redis: ping failed: %w", op, err)
	}

	return client, nil
}
