package strategy

import (
	"errors"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// 간단한 fake RedisClient 구현 (이 테스트에서 필요한 부분만)
type FakeRedisClient struct {
	mu         sync.Mutex
	store      map[string]int64
	errorOnKey string // 이 키에 대해서는 Incr 시 에러를 리턴
}

func NewFakeRedisClient() *FakeRedisClient {
	return &FakeRedisClient{
		store: make(map[string]int64),
	}
}

func (f *FakeRedisClient) Get(key string) (interface{}, error)                           { return nil, nil }
func (f *FakeRedisClient) Set(key string, value interface{}, expiration int) error       { return nil }
func (f *FakeRedisClient) GetObject(key string) (interface{}, error)                     { return nil, nil }
func (f *FakeRedisClient) SetObject(key string, value interface{}, expiration int) error { return nil }
func (f *FakeRedisClient) RemoveOldEntries(key string, cutoff time.Time) error           { return nil }
func (f *FakeRedisClient) AddToSortedSet(key, member string, score time.Time) error      { return nil }
func (f *FakeRedisClient) GetOldestEntry(key string) (redis.Z, error)                    { return redis.Z{}, nil }
func (f *FakeRedisClient) RemoveOldEntry(key string, before time.Time) error             { return nil }

func (f *FakeRedisClient) Incr(key string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errorOnKey != "" && key == f.errorOnKey {
		return 0, errors.New("simulated incr error")
	}
	f.store[key]++
	return f.store[key], nil
}

func (f *FakeRedisClient) Expire(key string, seconds int) {
	// 테스트에서는 TTL 무시
}

func (f *FakeRedisClient) ZRemRangeByScore(key string, from string, to string) error { return nil }
func (f *FakeRedisClient) ZSetSize(key string) int                                   { return 0 }
func (f *FakeRedisClient) ZCount(key string, min string, max string) (int, error) {
	return 0, nil
}
func (f *FakeRedisClient) HGetObject(key string) (interface{}, error) { return nil, nil }
func (f *FakeRedisClient) HSetObject(key string, value interface{}, expiration int) error {
	return nil
}

// helper: 간단한 ApiMatchResult 생성
func makeAPIMatchResult() *types.ApiMatchResult {
	return &types.ApiMatchResult{
		IsMatch:       true,
		Identifier:    "comment_write",
		Limit:         2,
		WindowSeconds: 2, // 짧게 잡아서 테스트 빠르게
		ExpireSeconds: 3600,
		Target:        "https://example.com",
	}
}

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
	redisClient := NewFakeRedisClient()
	limiter := NewFixedWindowCounterLimiter(keyGen, redisClient, cfg)

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
	apiResult := makeAPIMatchResult()
	// key generator uses matching strategy so key strings align
	keyGen := util.NewIpKeyGenerator(settings.RateLimiterConfig{Strategy: "fixed_window_counter"})

	t.Run("allowed until limit then blocked", func(t *testing.T) {
		fakeRedis := NewFakeRedisClient()
		limiterIface := NewFixedWindowCounterLimiter(keyGen, fakeRedis, settings.RateLimiterConfig{})
		// underlying struct assertion for access (optional)
		fwc := limiterIface.(*FixedWindowCounterLimiter)

		// first request: allowed
		decision1 := fwc.IsAllowed("127.0.0.1", apiResult, nil)
		if !decision1.Allowed {
			t.Fatalf("expected first request allowed, got %+v", decision1)
		}
		if decision1.Remaining != apiResult.Limit-1 {
			t.Errorf("expected remaining %d, got %d", apiResult.Limit-1, decision1.Remaining)
		}

		// second request: allowed, remaining should be zero
		decision2 := fwc.IsAllowed("127.0.0.1", apiResult, nil)
		if !decision2.Allowed {
			t.Fatalf("expected second request allowed, got %+v", decision2)
		}
		if decision2.Remaining != 0 {
			t.Errorf("expected remaining 0, got %d", decision2.Remaining)
		}

		// third request: over limit -> blocked
		decision3 := fwc.IsAllowed("127.0.0.1", apiResult, nil)
		if decision3.Allowed {
			t.Fatalf("expected third request denied, got %+v", decision3)
		}
		if decision3.Remaining != 0 {
			t.Errorf("expected remaining 0 when blocked, got %d", decision3.Remaining)
		}
		if decision3.RetryAfterSec < 0 {
			t.Errorf("expected non-negative RetryAfterSec, got %d", decision3.RetryAfterSec)
		}
	})

	t.Run("redis error causes denial", func(t *testing.T) {
		fakeRedis := NewFakeRedisClient()
		// 특정 키에 에러를 내도록 설정
		key := util.NewIpKeyGenerator(settings.RateLimiterConfig{Strategy: "fixed_window_counter"}).Make("127.0.0.1", apiResult.Identifier)
		fakeRedis.errorOnKey = key

		limiterIface := NewFixedWindowCounterLimiter(keyGen, fakeRedis, settings.RateLimiterConfig{})
		fwc := limiterIface.(*FixedWindowCounterLimiter)

		decision := fwc.IsAllowed("127.0.0.1", apiResult, nil)
		if decision.Allowed {
			t.Errorf("expected denial when redis Incr fails, got allowed")
		}
	})
}
