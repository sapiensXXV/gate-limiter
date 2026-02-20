package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log/slog"
)

type LeakyBucketLimiter struct {
	Config  settings.RateLimiterConfig
	Manager *LeakyBucketManager
}

var _ types.RateLimiter = (*LeakyBucketLimiter)(nil)

func NewLeakyBucketLimiter(
	config settings.RateLimiterConfig,
	manager *LeakyBucketManager,
) types.RateLimiter {
	h := &LeakyBucketLimiter{}
	h.Config = config
	h.Manager = manager
	return h
}

func (l *LeakyBucketLimiter) IsTarget(method, requestPath string) *types.ApiMatchResult {
	for _, api := range l.Config.Apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var isPathMatch bool
		if pathExpression == regex {
			isPathMatch = util.MatchRegex(requestPath, targetPath)
		} else if pathExpression == plain {
			isPathMatch = util.MatchPlain(requestPath, targetPath)
		} else {
			slog.Warn("unknown path expression", "expression", pathExpression)
		}
		if isPathMatch && method == api.Method {
			return &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    api.Identifier,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				ExpireSeconds: api.ExpireSeconds,
				RefillSeconds: api.RefillSeconds,
				Target:        api.Target,
			}
		}
	}
	return &types.ApiMatchResult{IsMatch: false}
}

func (l *LeakyBucketLimiter) IsAllowed(
	ip string,
	api *types.ApiMatchResult,
	queuedRequest *types.QueuedRequest,
) types.RateLimitDecision {
	item, enqueued := l.Manager.Enqueue(api.Identifier, ip, *api)

	freeSpace, err := l.Manager.CountBucketFreeCapacity(api.Identifier, ip)
	if err != nil {
		slog.Error("cannot check bucket capacity", "error", err)
		freeSpace = 0
	}

	retryAfterSec, err := l.Manager.CalcRetryTimeAfter(api.Identifier, ip, *api)
	if err != nil {
		slog.Error("cannot calc retry time", "error", err)
		retryAfterSec = 0
	}

	if !enqueued {
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     freeSpace,
			RetryAfterSec: retryAfterSec,
		}
	}

	if queuedRequest != nil && queuedRequest.Request != nil {
		select {
		case <-item.Done:
			// ok
		case <-queuedRequest.Request.Context().Done():
			return types.RateLimitDecision{
				Allowed:       false,
				Remaining:     freeSpace,
				RetryAfterSec: 0,
			}
		}
	} else {
		<-item.Done
	}

	return types.RateLimitDecision{
		Allowed:       true,
		Remaining:     freeSpace,
		RetryAfterSec: 0,
	}
}
