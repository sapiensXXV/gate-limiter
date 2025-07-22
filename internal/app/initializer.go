package app

import (
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
)

func InitializeRateHandler() *limiter.RateLimitHandler {

	redisClient := initRedisClient()
	keyGenerator := initKeyGenerator()

	responder := limiter.NewHttpLimitResponder(redisClient, keyGenerator)
	proxy := limiter.NewDefaultProxyHandler()
	matcher := limiter.NewHttpRateLimitMatcher(keyGenerator)

	return limiter.NewRateLimitHandler(matcher, proxy, responder)
}

func initKeyGenerator() *limiter.IpKeyGenerator {
	return &limiter.IpKeyGenerator{}
}

func initRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
