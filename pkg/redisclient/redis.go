package redisclient

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
	"time"
)

var (
	ctx = context.Background()
	Rdb *redis.Client
)

func InitRedis() {
	db_value, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		panic(err)
	}
	Rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password
		DB:       db_value,
	})

	if err := Rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("redis connection fail", err)
	} else {
		fmt.Println("redis connection success")
	}
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return Rdb.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (interface{}, error) {
	return Rdb.Get(ctx, key).Result()
}
