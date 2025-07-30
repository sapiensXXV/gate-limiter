package types

import (
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, expiration int) error
	GetObject(key string) (interface{}, error)
	SetObject(key string, value interface{}, expiration int) error
	RemoveOldEntries(key string, cutoff time.Time) error
	AddToSortedSet(key, member string, score time.Time) error
	GetOldestEntry(key string) (redis.Z, error)
	RemoveOldEntry(key string, before time.Time) error

	Incr(key string) (int64, error)
	Expire(key string, seconds int)

	ZRemRangeByScore(key string, from string, to string) error
	ZSetSize(key string) int
	ZCount(key string, min string, max string) (int, error)

	HGetObject(key string) (interface{}, error)
	HSetObject(key string, value interface{}, expiration int) error
}
