package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
)

const (
	regex = "regex"
	plain = "plain"
)

var strategies = map[string]func(keyGenerator limiter.KeyGenerator, redisClient redisclient.RedisClient, config config_ratelimiter.RateLimiterConfig) RateLimiter{
	"sliding_window_log": NewSlidingWindowLogLimiter,
}
