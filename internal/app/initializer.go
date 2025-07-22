package app

import (
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
)

func InitializeRateHandler() *limiter.RateLimitHandler {

	redisClient := initRedisClient()

	responder := limiter.NewHttpLimitResponder(redisClient)
	proxy := limiter.NewDefaultProxyHandler()
	matcher := limiter.NewHttpRateLimitMatcher()

	return limiter.NewRateLimitHandler(matcher, proxy, responder)
}

func initRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
