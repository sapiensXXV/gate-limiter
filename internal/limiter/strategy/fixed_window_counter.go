package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
)

type FixedWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*FixedWindowCounterLimiter)(nil)

func NewFixedWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	redisClient types.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &FixedWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (f *FixedWindowCounterLimiter) IsTarget(method, url string) *types.ApiMatchResult {
	//TODO implement me
	panic("implement me")
}

func (f *FixedWindowCounterLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) (bool, int) {
	//TODO implement me
	panic("implement me")
}
