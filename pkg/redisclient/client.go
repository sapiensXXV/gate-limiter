package redisclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bucket2 "gate-limiter/internal/limiter/bucket"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

type RedisClient interface {
	Get(key string) (interface{}, error)
	GetObject(key string) (interface{}, error)
	Set(key string, value interface{}, expiration int) error
	SetObject(key string, value interface{}, expiration int) error
	RemoveOldEntries(key string, cutoff time.Time) error
	AddToSortedSet(key, member string, score time.Time) error
	GetZSetSize(key string) int
	GetOldestEntry(key string) (redis.Z, error)
	RemoveOldEntry(key string, before time.Time) error
}

type DefaultRedisClient struct {
	ctx    context.Context
	client *redis.Client
}

var _ RedisClient = (*DefaultRedisClient)(nil)

func NewDefaultRedisClient() *DefaultRedisClient {
	dbValue, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("redis initialization fail")
	}
	rc := &DefaultRedisClient{}
	rc.ctx = context.Background()
	rc.client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password
		DB:       dbValue,
	})

	if err := rc.client.Ping(rc.ctx).Err(); err != nil {
		log.Fatal("redis connection fail")
	}
	log.Println("redis connection success")

	return rc
}

func (d *DefaultRedisClient) Get(key string) (interface{}, error) {
	val, err := d.client.Get(d.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		log.Printf("redis key=[%s] not exists\n", key)
		return nil, nil
	}
	return val, err
}

// GetObject 주어진 키에 해당하는 JSON 문자열을 역직렬화하여 객체를 반환합니다.
func (d *DefaultRedisClient) GetObject(key string) (interface{}, error) {
	val, err := d.client.Get(d.ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		log.Printf("redis key=[%s] not exists\n", key)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Redis GetObject error for key=[%s]: %w\n", key, err)
	}

	var bucket bucket2.TokenBucket
	err = json.Unmarshal(val, &bucket)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshal error for key=[%s]: %w\n", key, err)
	}
	return &bucket, nil
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

func (d *DefaultRedisClient) GetZSetSize(key string) int {
	size, err := d.client.ZCard(ctx, key).Result()
	if err != nil {
		log.Println("redis: get zset size fail")
	}
	return int(size)
}

func (d *DefaultRedisClient) GetOldestEntry(key string) (redis.Z, error) {
	vals, err := d.client.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		log.Println("redis: get oldest entry fail")
	}
	return vals[0], err
}

// zset

func (d *DefaultRedisClient) RemoveOldEntry(key string, before time.Time) error {
	score := float64(before.Unix())
	return d.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", score)).Err()
}
