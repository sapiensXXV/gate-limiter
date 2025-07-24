package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
)

const (
	regex = "regex"
	plain = "plain"
)

var Strategies = map[string]func(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter{
	"token_bucket":           NewTokenBucketLimiter,
	"leaky_bucket":           NewLeakyBucketLimiter,
	"fixed_window_counter":   NewFixedWindowCounterLimiter,
	"sliding_window_log":     NewSlidingWindowLogLimiter,
	"sliding_window_counter": NewSlidingWindowCounterLimiter,
}
