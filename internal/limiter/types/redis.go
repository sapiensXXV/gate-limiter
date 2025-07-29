package types

import (
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient interface {
	Get(key string) (interface{}, error)
	GetObject(key string) (interface{}, error)
	HGetObject(key string) (interface{}, error)
	Set(key string, value interface{}, expiration int) error
	SetObject(key string, value interface{}, expiration int) error
	HSetObject(key string, value interface{}, expiration int) error
	RemoveOldEntries(key string, cutoff time.Time) error
	AddToSortedSet(key, member string, score time.Time) error
	GetZSetSize(key string) int
	GetOldestEntry(key string) (redis.Z, error)
	RemoveOldEntry(key string, before time.Time) error
}
