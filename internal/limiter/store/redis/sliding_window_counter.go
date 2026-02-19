package redis

import (
	"context"
	"gate-limiter/internal/limiter/store"

	goredis "github.com/redis/go-redis/v9"
)

// Atomic Lua script for sliding window counter.
// Performs cleanup, weighted count calculation, conditional ZADD all atomically.
// Fixes race condition: 5 non-atomic calls → single Lua script.
var slidingWindowCounterScript = goredis.NewScript(`
local key        = KEYS[1]
local limit      = tonumber(ARGV[1])
local window_sec = tonumber(ARGV[2])
local now_unix   = tonumber(ARGV[3])
local member     = ARGV[4]

local cutoff = now_unix - window_sec

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, '-inf', cutoff)

-- Compute current window start (truncate to window boundary)
local current_window_start = now_unix - (now_unix % window_sec)
local gap = now_unix - current_window_start

-- Total entries in sorted set
local total = redis.call('ZCARD', key)

-- Entries in the current window
local current_count = redis.call('ZCOUNT', key, current_window_start, '+inf')

-- Previous window count
local prev_count = total - current_count

-- Weighted count: current + prev * (remaining fraction of window)
local weight = (window_sec - gap) / window_sec
local weighted = current_count + prev_count * weight

if weighted >= limit then
    -- Denied: retry after current window ends
    local retry_after = (current_window_start + window_sec) - now_unix
    if retry_after < 0 then retry_after = 0 end
    return {0, 0, retry_after}
end

-- Allowed: add entry
redis.call('ZADD', key, now_unix, member)
redis.call('EXPIRE', key, window_sec * 2)

local remaining = limit - math.floor(weighted) - 1
if remaining < 0 then remaining = 0 end

return {1, remaining, 0}
`)

type RedisSlidingWindowCounterStore struct {
	client *goredis.Client
}

var _ store.SlidingWindowCounterStore = (*RedisSlidingWindowCounterStore)(nil)

func NewSlidingWindowCounterStore(client *goredis.Client) *RedisSlidingWindowCounterStore {
	return &RedisSlidingWindowCounterStore{client: client}
}

func (s *RedisSlidingWindowCounterStore) Allow(ctx context.Context, key string, limit, windowSec int, nowUnix int64, member string) (store.SlidingWindowResult, error) {
	result, err := slidingWindowCounterScript.Run(ctx, s.client, []string{key}, limit, windowSec, nowUnix, member).Int64Slice()
	if err != nil {
		return store.SlidingWindowResult{}, err
	}

	return store.SlidingWindowResult{
		Allowed:       result[0] == 1,
		Remaining:     int(result[1]),
		RetryAfterSec: int(result[2]),
	}, nil
}
