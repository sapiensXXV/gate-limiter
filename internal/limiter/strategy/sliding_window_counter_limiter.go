package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
)

type SlidingWindowCounterLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       settings.RateLimiterConfig
}

var _ RateLimiter = (*SlidingWindowCounterLimiter)(nil)

func NewSlidingWindowCounterLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config settings.RateLimiterConfig,
) RateLimiter {
	return &SlidingWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (s *SlidingWindowCounterLimiter) IsTarget(method, url string) (bool, *ApiMatchResult) {
	//TODO implement me
	panic("implement me")
}

func (s *SlidingWindowCounterLimiter) IsAllowed(ip string, api *ApiMatchResult, _ *QueuedRequest) (bool, int) {
	//TODO implement me
	panic("implement me")
}
