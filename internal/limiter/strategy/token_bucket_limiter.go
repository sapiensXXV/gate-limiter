package strategy

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/bucket"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/pkg/redisclient"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type TokenBucketLimiter struct {
	KeyGenerator limiterutil.KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
}

var _ RateLimiter = (*TokenBucketLimiter)(nil)

func NewTokenBucketLimiter(
	keyGenerator limiterutil.KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) RateLimiter {
	h := &TokenBucketLimiter{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (l *TokenBucketLimiter) IsTarget(method, requestPath string) (bool, *ApiMatchResult) {
	apis := l.Config.Apis
	for _, api := range apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var result bool
		if pathExpression == regex {
			result = limiterutil.MatchRegex(requestPath, targetPath)
		} else if pathExpression == plain {
			result = limiterutil.MatchPlain(requestPath, targetPath)
		}
		if result && method == api.Method {
			return true, &ApiMatchResult{
				Identifier:    api.Key,
				Limit:         api.Limit,
				WindowSeconds: api.WindowSeconds,
				RefillSeconds: api.RefillSeconds,
				ExpireSeconds: api.ExpireSeconds,
				Target:        api.Target,
			}
		}
	}
	return false, nil
}

func (l *TokenBucketLimiter) IsAllowed(ip string, api *ApiMatchResult) (bool, int) {
	key := l.KeyGenerator.Make(ip, api.Identifier)
	b, err := l.RedisClient.GetObject(key)
	bb, ok := b.(*bucket.TokenBucket)
	if !ok {
		log.Println("Invalid type assertion for key [%s]", key)
		return false, 0
	}
	if err == redis.Nil {
		// 버킷이 없는 경우 새로운 버킷을 만들고 마지막 토큰 리필 시간을 현재로 설정한다.
		newBucket := bucket.NewTokenBucket(api.Limit)
		newBucket.LastRefillTime = time.Now()
		newBucket.Token-- // 토큰 한개 소비
		if err := l.RedisClient.SetObject(key, newBucket, api.ExpireSeconds); err != nil {
			log.Printf("redis value setting error: key=[%s], value=[%s], err=%v", key, newBucket, err)
			return false, 0
		}
		return true, newBucket.Token

	} else if err != nil {
		log.Printf("redis Get error key:[%s]\n, err:%v\n", key, err)
		return false, 0
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
			return false, 0
		}
		return true, bb.Token
	}
	log.Printf("Not enough token: user=[%s]", key)
	return false, 0 // 토큰이 없는 경우 요청을 거부한다.
}

func refillTokenIfNeeded(b *bucket.TokenBucket, limit int, refillSeconds int) {
	if time.Since(b.LastRefillTime).Seconds() > float64(refillSeconds) {
		b.Token = limit
		b.LastRefillTime = time.Now()
	}
}
