package app

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
)

func InitializeRateHandler(config config_ratelimiter.RateLimiterConfig) *limiter.RateLimitHandler {

	redisClient := initRedisClient()
	keyGenerator := initKeyGenerator()

	responder := limiter.NewHttpLimitResponder(nil, redisClient, keyGenerator)
	proxy := limiter.NewDefaultProxyHandler()
	matcher := limiter.NewHttpRateLimitMatcher(keyGenerator, redisClient)

	return limiter.NewRateLimitHandler(matcher, proxy, responder, config)
}

func initKeyGenerator() *limiter.IpKeyGenerator {
	return &limiter.IpKeyGenerator{}
}

func initRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
