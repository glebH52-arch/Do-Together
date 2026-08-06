package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, addr, pasword string, db int) (*goredis.Client, error) {
	options := goredis.Options{
		Addr:     addr,
		Password: pasword,
		DB:       db,
	}
	client := goredis.NewClient(&options)

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
