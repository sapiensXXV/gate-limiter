package redisclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gate-limiter/internal/limiter/types"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

type DefaultRedisClient struct {
	ctx    context.Context
	client *redis.Client
}

var _ types.RedisClient = (*DefaultRedisClient)(nil)

func NewDefaultRedisClient() *DefaultRedisClient {
	dbValue, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("redisclient initialization fail")
	}
	rc := &DefaultRedisClient{}
	rc.ctx = context.Background()
	rc.client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password
		DB:       dbValue,
	})

	if err := rc.client.Ping(rc.ctx).Err(); err != nil {
		log.Fatal("redisclient connection fail")
	}
	log.Println("redisclient connection success")

	return rc
}

func (d *DefaultRedisClient) Get(key string) (interface{}, error) {
	val, err := d.client.Get(d.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		log.Printf("redisclient key=[%s] not exists\n", key)
		return nil, nil
	}
	return val, err
}

// GetObject 주어진 키에 해당하는 JSON 문자열을 역직렬화하여 객체를 반환합니다.
func (d *DefaultRedisClient) GetObject(key string) (interface{}, error) {
	val, err := d.client.Get(d.ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		log.Printf("redisclient key=[%s] not exists\n", key)
		return nil, redis.Nil
	}
	if err != nil {
		return nil, fmt.Errorf("Redis GetObject error for key=[%s]: %w\n", key, err)
	}

	var bucket types.TokenBucket
	err = json.Unmarshal(val, &bucket)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshal error for key=[%s]: %w\n", key, err)
	}
	return &bucket, nil
}

func (d *DefaultRedisClient) HGetObject(key string) (interface{}, error) {
	// TODO implement this method
	return nil, nil
}

func (d *DefaultRedisClient) Set(key string, value interface{}, expiration int) error {
	return d.client.Set(d.ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

// SetObject 주어진 객체를 JSON 문자열로 직렬화하여 Redis에 저장한다.
func (d *DefaultRedisClient) SetObject(key string, value interface{}, expiration int) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON marshal error for key=[%s]: %w\n", key, err)
	}
	return d.client.Set(d.ctx, key, jsonData, time.Duration(expiration)*time.Second).Err()
}

func (d *DefaultRedisClient) HSetObject(key string, value interface{}, expiration int) error {
	// TODO implement this method
	return errors.New("new error")
}

func (d *DefaultRedisClient) RemoveOldEntries(key string, cutoff time.Time) error {
	score := float64(cutoff.Unix())
	return d.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatFloat(score, 'f', -1, 64)).Err()
}

func (d *DefaultRedisClient) AddToSortedSet(key, member string, t time.Time) error {
	score := float64(t.Unix())
	z := &redis.Z{
		Score:  score,
		Member: member,
	}
	return d.client.ZAdd(ctx, key, *z).Err()
}

func (d *DefaultRedisClient) ZSetSize(key string) int {
	size, err := d.client.ZCard(ctx, key).Result()
	if err != nil {
		log.Println("redisclient: get zset size fail")
	}
	return int(size)
}

func (d *DefaultRedisClient) GetOldestEntry(key string) (redis.Z, error) {
	vals, err := d.client.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		log.Println("redisclient: get oldest entry fail")
	}
	return vals[0], err
}

// zset

func (d *DefaultRedisClient) RemoveOldEntry(key string, before time.Time) error {
	score := float64(before.Unix())
	return d.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", score)).Err()
}

func (d *DefaultRedisClient) Incr(key string) (int64, error) {
	return d.client.Incr(d.ctx, key).Result()
}

func (d *DefaultRedisClient) Expire(key string, seconds int) {
	d.client.Expire(d.ctx, key, time.Duration(seconds)*time.Second)
}

func (d *DefaultRedisClient) ZRemRangeByScore(key string, from string, to string) error {
	return d.client.ZRemRangeByScore(d.ctx, key, from, to).Err()
}

func (d *DefaultRedisClient) ZCount(key string, min string, max string) (int, error) {
	cnt, err := d.client.ZCount(ctx, key, min, max).Result()
	return int(cnt), err
}
