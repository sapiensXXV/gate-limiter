package redisclient

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
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
		fmt.Println("redisclient connection fail", err)
	} else {
		fmt.Println("redisclient connection success")
	}
}
