package strategy

import (
	"context"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/store"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log/slog"
	"math"
	"time"
)

type FixedWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	Store        store.CounterStore
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*FixedWindowCounterLimiter)(nil)

func NewFixedWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	counterStore store.CounterStore,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &FixedWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		Store:        counterStore,
		Config:       config,
	}
}

func (l *FixedWindowCounterLimiter) IsTarget(requestMethod, requestURL string) *types.ApiMatchResult {
	apis := l.Config.Apis
	for _, api := range apis {
		expressionType := api.Path.Expression
		pathValue := api.Path.Value
		var result bool
		if expressionType == regex {
			result = util.MatchRegex(requestURL, pathValue)
		} else if expressionType == plain {
			result = util.MatchPlain(requestURL, pathValue)
		} else {
			slog.Error("unknown expression type", "type", expressionType)
			continue
		}

		if result && api.Method == requestMethod {
			return &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    api.Identifier,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				ExpireSeconds: api.ExpireSeconds,
				Target:        api.Target,
			}
		}
	}

	return &types.ApiMatchResult{IsMatch: false}
}

func (l *FixedWindowCounterLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) types.RateLimitDecision {
	windowStart := time.Now().Truncate(time.Duration(api.WindowSeconds) * time.Second)
	key := l.KeyGenerator.Make(ip, api.Identifier)

	result, err := l.Store.IncrementAndGet(context.TODO(), key, api.WindowSeconds)
	if err != nil {
		slog.Error("counter store error", "error", err)
		return types.RateLimitDecision{Allowed: false}
	}

	if result.Count > int64(api.Limit) {
		retryAt := windowStart.Add(time.Duration(api.WindowSeconds) * time.Second)
		wait := retryAt.Sub(time.Now())
		sec := int(math.Ceil(wait.Seconds()))
		if sec < 0 {
			sec = 0
		}

		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: sec,
		}
	}

	return types.RateLimitDecision{
		Allowed:       true,
		Remaining:     api.Limit - int(result.Count),
		RetryAfterSec: 0,
	}
}
