package redis

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func NewClient(config *settings.RedisClientConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + strconv.Itoa(config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return client, nil
}
