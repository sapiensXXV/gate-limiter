package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
)

type SlidingWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*SlidingWindowCounterLimiter)(nil)

func NewSlidingWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	redisClient types.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &SlidingWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (s *SlidingWindowCounterLimiter) IsTarget(method, url string) *types.ApiMatchResult {
	//TODO implement me
	panic("implement me")
}

func (s *SlidingWindowCounterLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) (bool, int) {
	//TODO implement me
	panic("implement me")
}
