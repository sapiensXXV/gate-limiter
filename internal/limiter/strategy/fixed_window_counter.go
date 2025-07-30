package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
	"time"
)

type FixedWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*FixedWindowCounterLimiter)(nil)

func NewFixedWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	redisClient types.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &FixedWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
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
			log.Fatalf("Unknown expression type: %s", expressionType)
		}

		if result && api.Method == requestMethod {
			// 통과
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
	log.Printf("ip_address: [%s]를 검사합니다.", ip)
	// time.Time.Truncate(d) 메서드는 현재 시간을 d단위로 내림 처리 해주는 역할을 한다.
	windowStart := time.Now().Truncate(time.Duration(api.WindowSeconds) * time.Second)
	key := l.KeyGenerator.Make(ip, api.Identifier) // redis_key

	cnt, err := l.RedisClient.Incr(key)
	if err != nil {
		return types.RateLimitDecision{Allowed: false}
	}

	// 최초 생성 때만 만료시간 세팅
	if cnt == 1 {
		l.RedisClient.Expire(key, api.WindowSeconds)
	}

	if cnt > int64(api.Limit) {
		retryAt := windowStart.Add(time.Duration(api.WindowSeconds) * time.Second)
		wait := retryAt.Sub(time.Now())
		sec := int(wait)
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
		Remaining:     api.Limit - int(cnt),
		RetryAfterSec: 0,
	}
}
