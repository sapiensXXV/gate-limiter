package redisclient

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

type RedisClient interface {
	RemoveOldEntries(key string, cutoff time.Time) error
	AddToSortedSet(key, member string, score time.Time) error
	GetZSetSize(key string) int
	GetOldestEntry(key string) (redis.Z, error)
}

type DefaultRedisClient struct {
	ctx context.Context
	rdb *redis.Client
}

var _ RedisClient = (*DefaultRedisClient)(nil)

func NewDefaultRedisClient() *DefaultRedisClient {
	dbValue, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("redis initialization fail")
	}
	rc := &DefaultRedisClient{}
	rc.ctx = context.Background()
	rc.rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password
		DB:       dbValue,
	})

	if err := rc.rdb.Ping(rc.ctx).Err(); err != nil {
		log.Fatal("redis connection fail")
	}
	log.Println("redis connection success")

	return rc
}

func (d *DefaultRedisClient) RemoveOldEntries(key string, cutoff time.Time) error {
	score := float64(cutoff.Unix())
	return d.rdb.ZRemRangeByScore(ctx, key, "0", strconv.FormatFloat(score, 'f', -1, 64)).Err()
}

func (d *DefaultRedisClient) AddToSortedSet(key, member string, t time.Time) error {
	score := float64(t.Unix())
	z := &redis.Z{
		Score:  score,
		Member: member,
	}
	return d.rdb.ZAdd(ctx, key, *z).Err()
}

func (d *DefaultRedisClient) GetZSetSize(key string) int {
	size, err := d.rdb.ZCard(ctx, key).Result()
	if err != nil {
		log.Println("redis: get zset size fail")
	}
	return int(size)
}

func (d *DefaultRedisClient) GetOldestEntry(key string) (redis.Z, error) {
	vals, err := d.rdb.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		log.Println("redis: get oldest entry fail")
	}
	return vals[0], err
}
