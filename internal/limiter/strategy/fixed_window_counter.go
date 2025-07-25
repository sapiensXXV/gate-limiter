package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
)

type FixedWindowCounterLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*FixedWindowCounterLimiter)(nil)

func NewFixedWindowCounterLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	return &FixedWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (f *FixedWindowCounterLimiter) IsTarget(method, url string) (bool, *HttpMatchResult) {
	//TODO implement me
	panic("implement me")
}

func (f *FixedWindowCounterLimiter) IsAllowed(ip string, api *HttpMatchResult) (bool, int) {
	//TODO implement me
	panic("implement me")
}
