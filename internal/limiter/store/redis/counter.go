package redis

import (
	"context"
	"gate-limiter/internal/limiter/store"

	"github.com/redis/go-redis/v9"
)

var counterScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if tonumber(current) == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`)

type RedisCounterStore struct {
	client *redis.Client
}

var _ store.CounterStore = (*RedisCounterStore)(nil)

func NewCounterStore(client *redis.Client) *RedisCounterStore {
	return &RedisCounterStore{client: client}
}

func (s *RedisCounterStore) IncrementAndGet(ctx context.Context, key string, windowSeconds int) (store.CounterResult, error) {
	result, err := counterScript.Run(ctx, s.client, []string{key}, windowSeconds).Int64()
	if err != nil {
		return store.CounterResult{}, err
	}
	return store.CounterResult{Count: result}, nil
}
