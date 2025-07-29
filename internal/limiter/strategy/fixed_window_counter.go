package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"gate-limiter/pkg/redisclient"
)

type FixedWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*FixedWindowCounterLimiter)(nil)

func NewFixedWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	redisClient redisclient.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &FixedWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (f *FixedWindowCounterLimiter) IsTarget(method, url string) (bool, *types.ApiMatchResult) {
	//TODO implement me
	panic("implement me")
}

func (f *FixedWindowCounterLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) (bool, int) {
	//TODO implement me
	panic("implement me")
}
