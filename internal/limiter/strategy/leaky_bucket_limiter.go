package strategy

import (
	config_ratelimiter "gate-limiter/config/limiterconfig"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
	"log"
)

type LeakyBucketLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
	Manager      *LeakyBucketManager
}

var _ RateLimiter = (*LeakyBucketLimiter)(nil)

func NewLeakyBucketLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
	manager *LeakyBucketManager,
) RateLimiter {
	h := &LeakyBucketLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	h.Manager = manager
	return h
}

func (l *LeakyBucketLimiter) IsTarget(method, url string) (bool, *ApiMatchResult) {
	for _, api := range l.Config.Apis {
		pathExpression := api.Path.Expression
		requestPath := api.Path.Value
		var isPathMatch bool
		if pathExpression == regex {
			isPathMatch = limiterutil.MatchRegex(url, requestPath)
		} else if pathExpression == plain {
			isPathMatch = limiterutil.MatchPlain(url, requestPath)
		} else {
			log.Println("cannot identify path expression")
		}
		if isPathMatch && method == api.Method {
			return true, &ApiMatchResult{
				Identifier: api.Identifier,
				Limit:      api.Limit,
				BucketSize: api.BucketSize,
				Target:     api.Target,
			}
		}
	}
	return false, nil
}

func (l *LeakyBucketLimiter) IsAllowed(ip string, api *ApiMatchResult, queuedRequest *QueuedRequest) (bool, int) {
	key := l.KeyGenerator.Make(ip, api.Identifier)
	result := l.Manager.AddRequest(key, *queuedRequest)
	// 큐에 여유공간이 있는지 확인하는 작업이 여기서는 채널에 데이터를 넣을 수 있는지 여부에 따라 결정된다.
	// 그 결과가 result 로 반환된다.
	freeSpace, err := l.Manager.CountBucketFreeCapacity(key)
	if err != nil {
		log.Println("Cannot check free space of channel", err)
	}
	return result, freeSpace
}
