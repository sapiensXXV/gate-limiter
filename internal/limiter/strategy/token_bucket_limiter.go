package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
	"time"
)

type TokenBucketLimiter struct {
	KeyGenerator util.KeyGenerator
	RedisClient  types.RedisClient
	Config       settings.RateLimiterConfig
}

var _ types.RateLimiter = (*TokenBucketLimiter)(nil)

// tokenBucketLuaScript 는 토큰 버킷의 조회·리필·소비를 원자적으로 처리하는 Lua 스크립트이다.
// GET→수정→SET 패턴의 Race Condition 을 방지한다.
const tokenBucketLuaScript = `
local key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local refill_sec = tonumber(ARGV[2])
local expire_sec = tonumber(ARGV[3])
local now_sec    = tonumber(ARGV[4])

local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens      = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens      = max_tokens
    last_refill = now_sec
end

if (now_sec - last_refill) >= refill_sec then
    tokens      = max_tokens
    last_refill = now_sec
end

local retry_after = (last_refill + refill_sec) - now_sec
if retry_after < 0 then retry_after = 0 end

if tokens > 0 then
    tokens = tokens - 1
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
    redis.call('EXPIRE', key, expire_sec)
    return {1, tokens, retry_after}
else
    return {0, 0, retry_after}
end
`

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
	now := time.Now().Unix()

	result, err := l.RedisClient.Eval(
		tokenBucketLuaScript,
		[]string{key},
		api.Limit,
		api.RefillSeconds,
		api.ExpireSeconds,
		now,
	)
	if err != nil {
		log.Printf("token bucket eval error: key=[%s], err=%v", key, err)
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: 0,
		}
	}

	vals, ok := result.([]interface{})
	if !ok || len(vals) < 3 {
		log.Printf("token bucket: unexpected result format: %v", result)
		return types.RateLimitDecision{
			Allowed:       false,
			Remaining:     0,
			RetryAfterSec: 0,
		}
	}

	allowed := vals[0].(int64) == 1
	remaining := int(vals[1].(int64))
	retryAfter := int(vals[2].(int64))

	return types.RateLimitDecision{
		Allowed:       allowed,
		Remaining:     remaining,
		RetryAfterSec: retryAfter,
	}
}
