package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
	"math"
	"time"
)

type SlidingWindowLogLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*SlidingWindowLogLimiter)(nil)

func NewSlidingWindowLogLimiter(
	keyGenerator util.KeyGenerator,
	redisClient types.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	h := &SlidingWindowLogLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (l *SlidingWindowLogLimiter) IsTarget(requestMethod, requestPath string) *types.ApiMatchResult {
	// 경로와 HTTP 메서드가 둘다 일치해야 제한 대상으로 판명
	apis := l.Config.Apis
	for _, api := range apis {
		expressionType := api.Path.Expression
		pathValue := api.Path.Value
		var result bool
		// 경로 표현 방식에 따라 경로 매칭 방식 결정
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
	size := l.RedisClient.ZSetSize(key)
	if size > api.Limit {
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: 0,
		}
	}

	return types.RateLimitDecision{
		Allowed:       true,
		Remaining:     api.Limit - size,
		RetryAfterSec: l.calcRetryAfterSeconds(key, api.WindowSeconds),
	}
}

func (l *SlidingWindowLogLimiter) calcRetryAfterSeconds(key string, windowSeconds int) int {
	oldest, err := l.RedisClient.GetOldestEntry(key)
	if err != nil {
		return 0
	}

	oldestSec := int64(math.Round(oldest.Score))
	oldestTime := time.Unix(oldestSec, 0)

	retryAt := oldestTime.Add(time.Duration(windowSeconds) * time.Second)
	wait := time.Until(retryAt).Seconds()

	if wait <= 0 {
		return 0
	}
	return int(math.Ceil(wait))

}
