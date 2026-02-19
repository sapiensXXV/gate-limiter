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

type SlidingWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	Store        store.SlidingWindowCounterStore
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*SlidingWindowCounterLimiter)(nil)

func NewSlidingWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	slidingWindowCounterStore store.SlidingWindowCounterStore,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &SlidingWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		Store:        slidingWindowCounterStore,
		Config:       config,
	}
}

func (l *SlidingWindowCounterLimiter) IsTarget(requestMethod, requestPath string) *types.ApiMatchResult {
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

func (l *SlidingWindowCounterLimiter) IsAllowed(
	ip string,
	api *types.ApiMatchResult,
	_ *types.QueuedRequest,
) types.RateLimitDecision {
	now := time.Now()
	key := l.KeyGenerator.Make(ip, api.Identifier)

	result, err := l.Store.Allow(context.TODO(), key, api.Limit, api.WindowSeconds, now.Unix(), now.String())
	if err != nil {
		log.Printf("sliding window counter store error: key=[%s], err=%v", key, err)
		return types.RateLimitDecision{Allowed: false}
	}

	return types.RateLimitDecision{
		Allowed:       result.Allowed,
		Remaining:     result.Remaining,
		RetryAfterSec: result.RetryAfterSec,
	}
}
