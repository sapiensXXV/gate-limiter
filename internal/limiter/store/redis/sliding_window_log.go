package redis

import (
	"context"
	"gate-limiter/internal/limiter/store"

	goredis "github.com/redis/go-redis/v9"
)

// Atomic Lua script: ZREMRANGEBYSCORE → ZCARD → conditional ZADD
// Fixes race condition: rejected requests are NOT added to the sorted set.
// Fixes panic on empty ZRANGE result by handling it safely inside Lua.
var slidingWindowLogScript = goredis.NewScript(`
local key        = KEYS[1]
local limit      = tonumber(ARGV[1])
local window_sec = tonumber(ARGV[2])
local now_unix   = tonumber(ARGV[3])
local member     = ARGV[4]

local cutoff = now_unix - window_sec

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, '-inf', cutoff)

-- Count current entries
local count = redis.call('ZCARD', key)

if count >= limit then
    -- Denied: compute retry_after from oldest entry
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local retry_after = 0
    if #oldest >= 2 then
        local oldest_score = tonumber(oldest[2])
        retry_after = (oldest_score + window_sec) - now_unix
        if retry_after < 0 then retry_after = 0 end
    end
    return {0, 0, retry_after}
end

-- Allowed: add entry
redis.call('ZADD', key, now_unix, member)
redis.call('EXPIRE', key, window_sec + 1)

local remaining = limit - count - 1

-- Compute retry_after from oldest entry
local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
local retry_after = 0
if #oldest >= 2 then
    local oldest_score = tonumber(oldest[2])
    retry_after = (oldest_score + window_sec) - now_unix
    if retry_after < 0 then retry_after = 0 end
end

return {1, remaining, retry_after}
`)

type RedisSlidingWindowLogStore struct {
	client *goredis.Client
}

var _ store.SlidingWindowLogStore = (*RedisSlidingWindowLogStore)(nil)

func NewSlidingWindowLogStore(client *goredis.Client) *RedisSlidingWindowLogStore {
	return &RedisSlidingWindowLogStore{client: client}
}

func (s *RedisSlidingWindowLogStore) Allow(ctx context.Context, key string, limit, windowSec int, nowUnix int64, member string) (store.SlidingWindowResult, error) {
	result, err := slidingWindowLogScript.Run(ctx, s.client, []string{key}, limit, windowSec, nowUnix, member).Int64Slice()
	if err != nil {
		return store.SlidingWindowResult{}, err
	}

	return store.SlidingWindowResult{
		Allowed:       result[0] == 1,
		Remaining:     int(result[1]),
		RetryAfterSec: int(result[2]),
	}, nil
}
