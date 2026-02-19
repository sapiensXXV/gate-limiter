package redis

import (
	"context"
	"gate-limiter/internal/limiter/store"
	"time"

	"github.com/redis/go-redis/v9"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local refill_sec = tonumber(ARGV[2])
local expire_sec = tonumber(ARGV[3])
local now_sec    = tonumber(ARGV[4])

local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens      = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens      = max_tokens
    last_refill = now_sec
end

if (now_sec - last_refill) >= refill_sec then
    tokens      = max_tokens
    last_refill = now_sec
end

local retry_after = (last_refill + refill_sec) - now_sec
if retry_after < 0 then retry_after = 0 end

if tokens > 0 then
    tokens = tokens - 1
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
    redis.call('EXPIRE', key, expire_sec)
    return {1, tokens, retry_after}
else
    return {0, 0, retry_after}
end
`)

type RedisTokenBucketStore struct {
	client *redis.Client
}

var _ store.TokenBucketStore = (*RedisTokenBucketStore)(nil)

func NewTokenBucketStore(client *redis.Client) *RedisTokenBucketStore {
	return &RedisTokenBucketStore{client: client}
}

func (s *RedisTokenBucketStore) TryConsume(ctx context.Context, key string, maxTokens, refillSec, expireSec int) (store.TokenBucketResult, error) {
	result, err := tokenBucketScript.Run(ctx, s.client, []string{key}, maxTokens, refillSec, expireSec, time.Now().Unix()).Slice()
	if err != nil {
		return store.TokenBucketResult{}, err
	}

	allowed := result[0].(int64) == 1
	remaining := int(result[1].(int64))
	retryAfter := int(result[2].(int64))

	return store.TokenBucketResult{
		Allowed:       allowed,
		Remaining:     remaining,
		RetryAfterSec: retryAfter,
	}, nil
}
