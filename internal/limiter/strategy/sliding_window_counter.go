package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
)

type SlidingWindowCounterLimiter struct {
	KeyGenerator limiter.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*SlidingWindowCounterLimiter)(nil)

func NewSlidingWindowCounterLimiter(
	keyGenerator limiter.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	return &SlidingWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (s *SlidingWindowCounterLimiter) IsTarget(method, url string) (bool, *config_ratelimiter.Api) {
	//TODO implement me
	panic("implement me")
}

func (s *SlidingWindowCounterLimiter) IsAllowed(ip string, api *config_ratelimiter.Api) (bool, int) {
	//TODO implement me
	panic("implement me")
}
