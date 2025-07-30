package strategy

import (
	"errors"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type TokenBucketLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*TokenBucketLimiter)(nil)

func NewTokenBucketLimiter(
	keyGenerator util.KeyGenerator,
	redisClient types.RedisClient,
	config settings.RateLimiterConfig,
) types.RateLimiter {
	h := &TokenBucketLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (l *TokenBucketLimiter) IsTarget(method, requestPath string) *types.ApiMatchResult {
	apis := l.Config.Apis
	for _, api := range apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var result bool
		if pathExpression == regex {
			result = util.MatchRegex(requestPath, targetPath)
		} else if pathExpression == plain {
			result = util.MatchPlain(requestPath, targetPath)
		}
		if result && method == api.Method {
			return &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    api.Identifier,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				RefillSeconds: api.RefillSeconds,
				ExpireSeconds: api.ExpireSeconds,
				Target:        api.Target,
			}
		}
	}
	return &types.ApiMatchResult{IsMatch: false}
}

func (l *TokenBucketLimiter) IsAllowed(ip string, api *types.ApiMatchResult, _ *types.QueuedRequest) types.RateLimitDecision {
	key := l.KeyGenerator.Make(ip, api.Identifier)
	b, err := l.RedisClient.GetObject(key)
	bb, ok := b.(*types.TokenBucket)
	if !ok {
		log.Println("Invalid type assertion for key [%s]", key)
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: 0,
		}
	}
	if errors.Is(err, redis.Nil) {
		// 버킷이 없는 경우 새로운 버킷을 만들고 마지막 토큰 리필 시간을 현재로 설정한다.
		newBucket := types.NewTokenBucket(api.Limit)
		newBucket.LastRefillTime = time.Now()
		newBucket.Token-- // 토큰 한개 소비
		if err := l.RedisClient.SetObject(key, newBucket, api.ExpireSeconds); err != nil {
			log.Printf("redisclient value setting error: key=[%s], value=[%s], err=%v", key, newBucket, err)
			return types.RateLimitDecision{
				Allowed:       false,
				Remaining:     0,
				RetryAfterSec: 0,
			}
		}
		return types.RateLimitDecision{
			Allowed:       true,
			Remaining:     newBucket.Token,
			RetryAfterSec: l.calcRetryAfter(newBucket, api), // TODO 재요청 가능시간 계산
		}
	} else if err != nil {
		log.Printf("redisclient Get error key:[%s]\n, err:%v\n", key, err)
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: 0,
		}
	}

	// 버킷이 있는 경우
	// 1. 마지막으로 버킷이 채워진 시간을 확인하고 토큰을 리필합니다.
	refillTokenIfNeeded(bb, api.Limit, api.RefillSeconds)

	// 토큰에 여유가 있는지 확인한다
	if bb.Token > 0 {
		bb.Token-- // 토큰을 한개 사용한다.
		err := l.RedisClient.SetObject(key, bb, api.ExpireSeconds)
		if err != nil {
			log.Printf("Redis SetObject Error:%v\n", err)
			return types.RateLimitDecision{
				Allowed:       false,
				Remaining:     0,
				RetryAfterSec: 0,
			}
		}
		return types.RateLimitDecision{
			Allowed:       true,
			Remaining:     bb.Token,
			RetryAfterSec: l.calcRetryAfter(bb, api),
		}
	}
	log.Printf("Not enough token: user=[%s]", key)
	return types.RateLimitDecision{
		Allowed:       false,
		Remaining:     0,
		RetryAfterSec: 0,
	}
}

func refillTokenIfNeeded(b *types.TokenBucket, limit int, refillSeconds int) {
	if time.Since(b.LastRefillTime).Seconds() > float64(refillSeconds) {
		b.Token = limit
		b.LastRefillTime = time.Now()
	}
}

func (l *TokenBucketLimiter) calcRetryAfter(b *types.TokenBucket, api *types.ApiMatchResult) int {
	nextRefillTime := b.LastRefillTime.Add(time.Duration(api.RefillSeconds) * time.Second)
	diff := nextRefillTime.Sub(time.Now())
	if diff <= 0 {
		return 0
	}
	return int(diff.Seconds())
}
