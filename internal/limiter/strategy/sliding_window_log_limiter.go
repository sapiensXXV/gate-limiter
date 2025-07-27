package strategy

import (
	"fmt"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
	"log"
	"time"
)

type SlidingWindowLogLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*SlidingWindowLogLimiter)(nil)

func NewSlidingWindowLogLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	h := &SlidingWindowLogLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (l *SlidingWindowLogLimiter) IsTarget(requestMethod, requestPath string) (bool, *ApiMatchResult) {
	// 경로와 HTTP 메서드가 둘다 일치해야 제한 대상으로 판명
	apis := l.Config.Apis
	for _, api := range apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var result bool
		// 경로 표현 방식에 따라 경로 매칭 방식 결정
		if pathExpression == regex {
			result = limiterutil.MatchRegex(requestPath, targetPath)
		} else if pathExpression == plain {
			result = limiterutil.MatchPlain(requestPath, targetPath)
		}
		if result && requestMethod == api.Method {
			return true, &ApiMatchResult{
				Identifier:    api.Key,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				Target:        api.Target,
			}
		}
	}
	return false, nil
}

func (l *SlidingWindowLogLimiter) IsAllowed(ip string, api *ApiMatchResult, _ *QueuedRequest) (bool, int) {
	fmt.Printf("ip_address: [%s]를 검사합니다.\n", ip)
	key := l.KeyGenerator.Make(ip, api.Identifier)

	var err error
	now := time.Now()

	err = l.RedisClient.RemoveOldEntries(key, now.Add(-time.Duration(api.WindowSeconds)*time.Second))
	if err != nil {
		log.Println("error while removing old entries:", err)
	}
	err = l.RedisClient.AddToSortedSet(key, now.String(), now)
	if err != nil {
		log.Println("error while adding to sorted set:", err)
	}
	size := l.RedisClient.GetZSetSize(key)
	if size > api.Limit {
		return false, 0
	}

	return true, api.Limit - size
}
