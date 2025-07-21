package app

import "gate-limiter/internal/limiter"

func InitializeRateHandler() *limiter.RateLimitHandler {

	responder := limiter.NewHttpLimitResponder()
	proxy := limiter.NewDefaultProxyHandler()
	matcher := limiter.NewHttpRateLimitMatcher()

	return limiter.NewRateLimitHandler(matcher, proxy, responder)
}
