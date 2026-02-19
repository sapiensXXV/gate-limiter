package strategy

import (
	"context"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/store"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
)

type TokenBucketLimiter struct {
	KeyGenerator util.KeyGenerator
	Store        store.TokenBucketStore
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*TokenBucketLimiter)(nil)

func NewTokenBucketLimiter(
	keyGenerator util.KeyGenerator,
	tokenBucketStore store.TokenBucketStore,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &TokenBucketLimiter{
		KeyGenerator: keyGenerator,
		Store:        tokenBucketStore,
		Config:       config,
	}
}

func (l *TokenBucketLimiter) IsTarget(method, requestPath string) *types.ApiMatchResult {
	apis := l.Config.Apis
	for _, api := range apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var result bool
		if pathExpression == regex {
			result = util.MatchRegex(requestPath, targetPath)
		} else if pathExpression == plain {
			result = util.MatchPlain(requestPath, targetPath)
		}
		if result && method == api.Method {
			return &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    api.Identifier,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				RefillSeconds: api.RefillSeconds,
				ExpireSeconds: api.ExpireSeconds,
				Target:        api.Target,
			}
		}
	}
	return &types.ApiMatchResult{IsMatch: false}
}

func (l *TokenBucketLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) types.RateLimitDecision {
	key := l.KeyGenerator.Make(ip, api.Identifier)

	result, err := l.Store.TryConsume(context.TODO(), key, api.Limit, api.RefillSeconds, api.ExpireSeconds)
	if err != nil {
		log.Printf("token bucket store error: key=[%s], err=%v", key, err)
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: 0,
		}
	}

	return types.RateLimitDecision{
		Allowed:       result.Allowed,
		Remaining:     result.Remaining,
		RetryAfterSec: result.RetryAfterSec,
	}
}
