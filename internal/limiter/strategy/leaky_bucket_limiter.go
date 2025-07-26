package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
	"log"
)

type LeakyBucketLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
	Manager      LeakyBucketManager
}

var _ RateLimiter = (*LeakyBucketLimiter)(nil)

func NewLeakyBucketLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	h := &LeakyBucketLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
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
				Identifier: api.Key,
				Limit:      api.Limit,
				BucketSize: api.BucketSize,
				Target:     api.Target,
			}
		}
	}
	return false, nil
}

func (l *LeakyBucketLimiter) IsAllowed(ip string, api *ApiMatchResult) (bool, int) {
	key := l.KeyGenerator.Make(ip, api.Identifier)
	l.Manager.AddRequest()
}
