package redisclient

import (
	"time"
)

func Set(key string, value interface{}, expiration time.Duration) error {
	return Rdb.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (interface{}, error) {
	return Rdb.Get(ctx, key).Result()
}
