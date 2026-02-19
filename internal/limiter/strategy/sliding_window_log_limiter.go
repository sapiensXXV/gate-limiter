package strategy

import (
	"context"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/store"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
	"time"
)

type SlidingWindowLogLimiter struct {
	KeyGenerator util.KeyGenerator
	Store        store.SlidingWindowLogStore
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*SlidingWindowLogLimiter)(nil)

func NewSlidingWindowLogLimiter(
	keyGenerator util.KeyGenerator,
	slidingWindowLogStore store.SlidingWindowLogStore,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &SlidingWindowLogLimiter{
		KeyGenerator: keyGenerator,
		Store:        slidingWindowLogStore,
		Config:       config,
	}
}

func (l *SlidingWindowLogLimiter) IsTarget(requestMethod, requestPath string) *types.ApiMatchResult {
	apis := l.Config.Apis
	for _, api := range apis {
		expressionType := api.Path.Expression
		pathValue := api.Path.Value
		var result bool
		if expressionType == regex {
			result = util.MatchRegex(requestPath, pathValue)
		} else if expressionType == plain {
			result = util.MatchPlain(requestPath, pathValue)
		}
		if result && requestMethod == api.Method {
			return &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    api.Identifier,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				Target:        api.Target,
			}
		}
	}
	return &types.ApiMatchResult{IsMatch: false}
}

func (l *SlidingWindowLogLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) types.RateLimitDecision {
	log.Printf("ip_address: [%s]를 검사합니다.\n", ip)
	key := l.KeyGenerator.Make(ip, api.Identifier)
	now := time.Now()

	result, err := l.Store.Allow(context.TODO(), key, api.Limit, api.WindowSeconds, now.Unix(), now.String())
	if err != nil {
		log.Printf("sliding window log store error: key=[%s], err=%v", key, err)
		return types.RateLimitDecision{Allowed: false}
	}

	return types.RateLimitDecision{
		Allowed:       result.Allowed,
		Remaining:     result.Remaining,
		RetryAfterSec: result.RetryAfterSec,
	}
}
