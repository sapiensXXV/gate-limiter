package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
	"math"
	"strconv"
	"time"
)

type SlidingWindowCounterLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*SlidingWindowCounterLimiter)(nil)

func NewSlidingWindowCounterLimiter(
	keyGenerator util.KeyGenerator,
	redisClient types.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	return &SlidingWindowCounterLimiter{
		KeyGenerator: keyGenerator,
		RedisClient:  redisClient,
		Config:       config,
	}
}

func (l *SlidingWindowCounterLimiter) IsTarget(requestMethod, requestPath string) *types.ApiMatchResult {
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

func (l *SlidingWindowCounterLimiter) IsAllowed(
	ip string,
	api *types.ApiMatchResult,
	_ *types.QueuedRequest,
) types.RateLimitDecision {
	now := time.Now()                               // 현재 시간
	cutoff := now.Unix() - int64(api.WindowSeconds) // 고려 대상 전의 시간
	key := l.KeyGenerator.Make(ip, api.Identifier)

	// 시간이 지난 데이터 삭제
	if err := l.RedisClient.RemoveOldEntries(key, now.Add(-time.Duration(api.WindowSeconds)*time.Second)); err != nil {
		log.Printf("fail to remove old entry on key=[%s] %v\n", key, err)
	}

	if err := l.RedisClient.ZRemRangeByScore(key, "0", strconv.FormatInt(cutoff, 10)); err != nil {
		log.Printf("fail to remove range by score, [key]:%s, from:[%s], to:[%d], err:%v", key, "0", cutoff, err)
		return types.RateLimitDecision{Allowed: false}
	}

	// 2) 새 요청 로그 추가
	if err := l.RedisClient.AddToSortedSet(key, now.String(), now); err != nil {
		log.Printf("fail to add to sorted set on key=[%s] %v\n", key, err)
		return types.RateLimitDecision{Allowed: false}
	}

	window := time.Duration(api.WindowSeconds) * time.Second
	currentWindowStart := now.Truncate(window)
	// currentWindowStart 시간으로부터 몇초가 지났는지를 봐야한다.
	gapFromCurrentStart := int(now.Sub(currentWindowStart).Seconds()) // 현재 시간으로부터 지난 초 수
	// 전체 카운트와, currentWindowStart 이후로 온 요청의 갯수를 센다면 비율을 파악할 수 있다.

	size := l.RedisClient.ZSetSize(key)
	searchMin := strconv.FormatInt(currentWindowStart.Unix(), 10)
	currentWindowSize, err := l.RedisClient.ZCount(key, searchMin, "+inf")
	if err != nil {
		log.Printf("fail to get current window size, [%s]\n", err)
		return types.RateLimitDecision{Allowed: false}
	}

	result := float64(currentWindowSize) + float64(size-currentWindowSize)*(float64(api.WindowSeconds-gapFromCurrentStart)/float64(api.WindowSeconds))
	if int(math.Floor(result)) > api.Limit {
		return types.RateLimitDecision{
			Allowed:       false,
			RetryAfterSec: int(currentWindowStart.Add(time.Duration(api.WindowSeconds) * time.Second).Sub(now)),
		}
	}

	return types.RateLimitDecision{Allowed: true}
}
