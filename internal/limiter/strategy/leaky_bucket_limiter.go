package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
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

func (l *LeakyBucketLimiter) IsTarget(method, url string) (bool, *types.ApiMatchResult) {
	for _, api := range l.Config.Apis {
		pathExpression := api.Path.Expression
		requestPath := api.Path.Value
		var isPathMatch bool
		if pathExpression == regex {
			isPathMatch = util.MatchRegex(url, requestPath)
		} else if pathExpression == plain {
			isPathMatch = util.MatchPlain(url, requestPath)
		} else {
			log.Println("cannot identify path expression")
		}
		if isPathMatch && method == api.Method {
			return true, &types.ApiMatchResult{
				Identifier: api.Identifier,
				Limit:      api.Limit,
				BucketSize: api.BucketSize,
				Target:     api.Target,
			}
		}
	}
	return false, nil
}

func (l *LeakyBucketLimiter) IsAllowed(
	ip string,
	api *types.ApiMatchResult,
	queuedRequest *types.QueuedRequest,
) (bool, int) {
	result := l.Manager.AddRequest(api.Identifier, ip, *queuedRequest, *api)
	// 큐에 여유공간이 있는지 확인하는 작업이 여기서는 채널에 데이터를 넣을 수 있는지 여부에 따라 결정된다.
	// 그 결과가 result 로 반환된다.
	freeSpace, err := l.Manager.CountBucketFreeCapacity(api.Identifier, ip)
	if err != nil {
		log.Println("Cannot check free space of channel", err)
	}
	return result, freeSpace
}
