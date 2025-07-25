package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
)

type LeakyBucketLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*LeakyBucketLimiter)(nil)

func NewLeakyBucketLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	h := &LeakyBucketLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (l *LeakyBucketLimiter) IsTarget(method, url string) (bool, *ApiMatchResult) {
	//TODO implement me
	panic("implement me")
}

func (l *LeakyBucketLimiter) IsAllowed(ip string, api *ApiMatchResult) (bool, int) {
	//TODO implement me
	panic("implement me")
}
