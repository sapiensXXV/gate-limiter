package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestFixedWindowCounterLimiter_IsTarget(t *testing.T) {
	cfg := settings.RateLimiterConfig{
		Strategy: "fixed_window_counter",
		Apis: []settings.Api{
			{
				Identifier:    "comment_write",
				Path:          settings.RateLimiterPath{Expression: plain, Value: "/api/item/123/comment"},
				Method:        "POST",
				Limit:         5,
				WindowSeconds: 60,
				ExpireSeconds: 3600,
				Target:        "http://target",
			},
			{
				Identifier:    "comment_regex",
				Path:          settings.RateLimiterPath{Expression: regex, Value: `^/api/item/\d+/comment$`},
				Method:        "POST",
				Limit:         3,
				WindowSeconds: 60,
				ExpireSeconds: 3600,
				Target:        "http://regex-target",
			},
		},
	}

	keyGen := util.NewIpKeyGenerator(settings.RateLimiterConfig{Strategy: "fixed_window_counter"})
	_, rc := newTestRedisClient(t)
	limiter := NewFixedWindowCounterLimiter(keyGen, rc, cfg)

	tests := []struct {
		name          string
		method        string
		url           string
		expectedMatch *types.ApiMatchResult
	}{
		{
			name:   "plain path matches",
			method: "POST",
			url:    "/api/item/123/comment",
			expectedMatch: &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    "comment_write",
				Limit:         5,
				WindowSeconds: 60,
				ExpireSeconds: 3600,
				Target:        "http://target",
			},
		},
		{
			name:   "regex path matches",
			method: "POST",
			url:    "/api/item/999/comment",
			expectedMatch: &types.ApiMatchResult{
				IsMatch:       true,
				Identifier:    "comment_regex",
				Limit:         3,
				WindowSeconds: 60,
				ExpireSeconds: 3600,
				Target:        "http://regex-target",
			},
		},
		{
			name:          "method mismatch",
			method:        "GET",
			url:           "/api/item/123/comment",
			expectedMatch: &types.ApiMatchResult{IsMatch: false},
		},
		{
			name:          "path mismatch",
			method:        "POST",
			url:           "/api/other",
			expectedMatch: &types.ApiMatchResult{IsMatch: false},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := limiter.IsTarget(tt.method, tt.url)
			if diff := cmp.Diff(tt.expectedMatch, got); diff != "" {
				t.Errorf("IsTarget() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFixedWindowCounterLimiter_IsAllowed(t *testing.T) {
	keyGen := util.NewIpKeyGenerator(settings.RateLimiterConfig{Strategy: "fixed_window_counter"})

	t.Run("allowed until limit then blocked", func(t *testing.T) {
		_, rc := newTestRedisClient(t)
		limiter := NewFixedWindowCounterLimiter(keyGen, rc, settings.RateLimiterConfig{})

		api := &types.ApiMatchResult{
			IsMatch:       true,
			Identifier:    "test_api",
			Limit:         2,
			WindowSeconds: 60,
			ExpireSeconds: 3600,
		}

		// 첫 번째 요청: 허용, remaining = 1
		d1 := limiter.IsAllowed("127.0.0.1", api, nil)
		assert.True(t, d1.Allowed)
		assert.Equal(t, 1, d1.Remaining)

		// 두 번째 요청: 허용, remaining = 0
		d2 := limiter.IsAllowed("127.0.0.1", api, nil)
		assert.True(t, d2.Allowed)
		assert.Equal(t, 0, d2.Remaining)

		// 세 번째 요청: 제한 초과 → 거부
		d3 := limiter.IsAllowed("127.0.0.1", api, nil)
		assert.False(t, d3.Allowed)
		assert.Equal(t, 0, d3.Remaining)
		assert.GreaterOrEqual(t, d3.RetryAfterSec, 0)
	})

	t.Run("different IPs have separate counters", func(t *testing.T) {
		_, rc := newTestRedisClient(t)
		limiter := NewFixedWindowCounterLimiter(keyGen, rc, settings.RateLimiterConfig{})

		api := &types.ApiMatchResult{
			IsMatch:       true,
			Identifier:    "test_api",
			Limit:         1,
			WindowSeconds: 60,
			ExpireSeconds: 3600,
		}

		d1 := limiter.IsAllowed("10.0.0.1", api, nil)
		assert.True(t, d1.Allowed)

		d2 := limiter.IsAllowed("10.0.0.2", api, nil)
		assert.True(t, d2.Allowed, "다른 IP는 별도 카운터를 가져야 한다")
	})
}
