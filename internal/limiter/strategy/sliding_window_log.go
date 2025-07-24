package strategy

import (
	"fmt"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
	"log"
	"regexp"
	"time"
)

type SlidingWindowLogLimiter struct {
	KeyGenerator limiter.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*SlidingWindowLogLimiter)(nil)

func NewSlidingWindowLogLimiter(
	keyGenerator limiter.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	h := &SlidingWindowLogLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (rc *SlidingWindowLogLimiter) IsTarget(requestMethod, requestPath string) (bool, *config_ratelimiter.Api) {
	// 경로와 HTTP 메서드가 둘다 일치해야 제한 대상으로 판명
	apis := rc.Config.Apis
	for _, api := range apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var result bool
		// 경로 표현 방식에 따라 경로 매칭 방식 결정
		if pathExpression == regex {
			result = rc.matchRegexPath(requestPath, targetPath)
		} else if pathExpression == plain {
			result = rc.matchPlainPath(requestPath, targetPath)
		}
		if result && requestMethod == api.Method {
			return true, &api
		}
	}
	return false, nil
}

func (rc *SlidingWindowLogLimiter) IsAllowed(ip string, api *config_ratelimiter.Api) (bool, int) {
	fmt.Printf("ip_adrress: [%s]를 검사합니다.\n", ip)
	key := rc.KeyGenerator.Make(ip, api.Key)

	var err error
	now := time.Now()

	err = rc.RedisClient.RemoveOldEntries(key, now.Add(-time.Duration(api.WindowSeconds)*time.Second))
	if err != nil {
		log.Println("error while removing old entries:", err)
	}
	err = rc.RedisClient.AddToSortedSet(key, now.String(), now)
	if err != nil {
		log.Println("error while adding to sorted set:", err)
	}
	size := rc.RedisClient.GetZSetSize(key)
	if size > api.Limit {
		return false, 0
	}

	return true, api.Limit - size
}

func (rc *SlidingWindowLogLimiter) matchPlainPath(requestPath string, target string) bool {
	return requestPath == target
}

func (rc *SlidingWindowLogLimiter) matchRegexPath(requestPath string, target string) bool {
	r, err := regexp.Compile(target)
	if err != nil {
		log.Println("error while compile regex:", err)
	}
	return r.MatchString(requestPath)
}
