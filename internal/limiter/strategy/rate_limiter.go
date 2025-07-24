package strategy

import "gate-limiter/config/ratelimiter"

type RateLimiter interface {
	IsTarget(method, url string) (bool, *config_ratelimiter.Api)
	IsAllowed(ip string, api *config_ratelimiter.Api) (bool, int)
}

type PathMatcher interface {
	Match(path string, target string) bool
}
