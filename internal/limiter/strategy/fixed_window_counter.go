package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
)

type FixedWindowCounterLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       settings.RateLimiterConfig
}

var _ RateLimiter = (*FixedWindowCounterLimiter)(nil)

func NewFixedWindowCounterLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config settings.RateLimiterConfig,
) RateLimiter {
	return &FixedWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (f *FixedWindowCounterLimiter) IsTarget(method, url string) (bool, *ApiMatchResult) {
	//TODO implement me
	panic("implement me")
}

func (f *FixedWindowCounterLimiter) IsAllowed(ip string, api *ApiMatchResult, _ *QueuedRequest) (bool, int) {
	//TODO implement me
	panic("implement me")
}
