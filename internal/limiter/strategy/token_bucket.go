package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
)

type TokenBucketLimiter struct {
	KeyGenerator limiter.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*TokenBucketLimiter)(nil)

func NewTokenBucketLimiter(
	keyGenerator limiter.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	h := &TokenBucketLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (t *TokenBucketLimiter) IsTarget(method, url string) (bool, *config_ratelimiter.Api) {
	//TODO implement me
	panic("implement me")
}

func (t *TokenBucketLimiter) IsAllowed(ip string, api *config_ratelimiter.Api) (bool, int) {
	//TODO implement me
	panic("implement me")
}
