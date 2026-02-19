package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenBucketLimiter_IsTarget(t *testing.T) {
	cfg := settings.RateLimiterConfig{
		Strategy: "token_bucket",
		Apis: []settings.Api{
			{
				Identifier:    "public_api",
				Path:          settings.RateLimiterPath{Expression: plain, Value: "/api/public/data"},
				Method:        "GET",
				Limit:         100,
				RefillSeconds: 10,
				WindowSeconds: 10,
				ExpireSeconds: 600,
			},
		},
	}

	keyGen := util.NewIpKeyGenerator(settings.RateLimiterConfig{Strategy: "token_bucket"})
	_, rc := newTestRedisClient(t)
	limiter := NewTokenBucketLimiter(keyGen, rc, cfg)

	t.Run("matching path and method", func(t *testing.T) {
		result := limiter.IsTarget("GET", "/api/public/data")
		assert.True(t, result.IsMatch)
		assert.Equal(t, "public_api", result.Identifier)
	})

	t.Run("non-matching path", func(t *testing.T) {
		result := limiter.IsTarget("GET", "/api/other")
		assert.False(t, result.IsMatch)
	})

	t.Run("non-matching method", func(t *testing.T) {
		result := limiter.IsTarget("POST", "/api/public/data")
		assert.False(t, result.IsMatch)
	})
}

func TestTokenBucketLimiter_IsAllowed(t *testing.T) {
	keyGen := util.NewIpKeyGenerator(settings.RateLimiterConfig{Strategy: "token_bucket"})

	t.Run("allowed until tokens exhausted", func(t *testing.T) {
		_, rc := newTestRedisClient(t)
		limiter := NewTokenBucketLimiter(keyGen, rc, settings.RateLimiterConfig{})

		api := &types.ApiMatchResult{
			IsMatch:       true,
			Identifier:    "test_api",
			Limit:         3,
			RefillSeconds: 60,
			ExpireSeconds: 600,
		}

		// 토큰 3개 소비: 모두 허용
		for i := 0; i < 3; i++ {
			d := limiter.IsAllowed("127.0.0.1", api, nil)
			assert.True(t, d.Allowed, "요청 %d는 허용되어야 한다", i+1)
			assert.Equal(t, 3-i-1, d.Remaining)
		}

		// 4번째 요청: 토큰 소진 → 거부
		d := limiter.IsAllowed("127.0.0.1", api, nil)
		assert.False(t, d.Allowed)
		assert.Equal(t, 0, d.Remaining)
	})

	t.Run("different IPs have separate buckets", func(t *testing.T) {
		_, rc := newTestRedisClient(t)
		limiter := NewTokenBucketLimiter(keyGen, rc, settings.RateLimiterConfig{})

		api := &types.ApiMatchResult{
			IsMatch:       true,
			Identifier:    "test_api",
			Limit:         1,
			RefillSeconds: 60,
			ExpireSeconds: 600,
		}

		d1 := limiter.IsAllowed("10.0.0.1", api, nil)
		assert.True(t, d1.Allowed)

		d2 := limiter.IsAllowed("10.0.0.2", api, nil)
		assert.True(t, d2.Allowed, "다른 IP는 별도 버킷을 가져야 한다")
	})

	t.Run("tokens refill after refill interval", func(t *testing.T) {
		mr, rc := newTestRedisClient(t)
		limiter := NewTokenBucketLimiter(keyGen, rc, settings.RateLimiterConfig{})

		api := &types.ApiMatchResult{
			IsMatch:       true,
			Identifier:    "refill_test",
			Limit:         2,
			RefillSeconds: 5,
			ExpireSeconds: 600,
		}

		// 토큰 2개 소진
		limiter.IsAllowed("127.0.0.1", api, nil)
		limiter.IsAllowed("127.0.0.1", api, nil)

		d := limiter.IsAllowed("127.0.0.1", api, nil)
		assert.False(t, d.Allowed, "토큰 소진 후 거부되어야 한다")

		// Redis의 last_refill을 과거로 조작하여 리필 트리거
		key := keyGen.Make("127.0.0.1", "refill_test")
		pastTime := time.Now().Unix() - 10
		mr.HSet(key, "last_refill", strconv.FormatInt(pastTime, 10))

		d = limiter.IsAllowed("127.0.0.1", api, nil)
		assert.True(t, d.Allowed, "리필 후 다시 허용되어야 한다")
		assert.Equal(t, 1, d.Remaining)
	})

	t.Run("retry_after is positive when denied", func(t *testing.T) {
		_, rc := newTestRedisClient(t)
		limiter := NewTokenBucketLimiter(keyGen, rc, settings.RateLimiterConfig{})

		api := &types.ApiMatchResult{
			IsMatch:       true,
			Identifier:    "retry_test",
			Limit:         1,
			RefillSeconds: 30,
			ExpireSeconds: 600,
		}

		limiter.IsAllowed("127.0.0.1", api, nil) // 토큰 소진

		d := limiter.IsAllowed("127.0.0.1", api, nil)
		assert.False(t, d.Allowed)
		assert.Greater(t, d.RetryAfterSec, 0, "거부 시 retry_after는 양수여야 한다")
	})
}
