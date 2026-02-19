package limiter

import (
	"gate-limiter/config/settings"
	storeredis "gate-limiter/internal/limiter/store/redis"
	"gate-limiter/internal/limiter/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestServeHTTP_NilLimiter(t *testing.T) {
	handler := &RateLimitHandler{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServeHTTP_NonMatchingRequest_Forwarded(t *testing.T) {
	proxy := &MockProxy{}
	handler := NewRateLimitHandler(
		&MockLimiter{IsTargetResult: &types.ApiMatchResult{IsMatch: false}},
		proxy,
		&MockResponder{},
		settings.RateLimiterConfig{
			Target: "http://example.com",
		},
		nil,
		&MockIdentifier{ClientID: "127.0.0.1"},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/unknown", nil)
	handler.ServeHTTP(w, r)

	assert.True(t, proxy.Called, "매칭되지 않는 요청은 origin으로 포워딩되어야 한다")
}

func TestServeHTTP_MatchingRequest_Allowed(t *testing.T) {
	proxy := &MockProxy{}
	responder := &MockResponder{}
	handler := NewRateLimitHandler(
		&MockLimiter{
			IsTargetResult:  &types.ApiMatchResult{IsMatch: true, Identifier: "test"},
			IsAllowedResult: types.RateLimitDecision{Allowed: true, Remaining: 5},
		},
		proxy,
		responder,
		settings.RateLimiterConfig{
			Target: "http://example.com",
		},
		nil,
		&MockIdentifier{ClientID: "127.0.0.1"},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/test", nil)
	handler.ServeHTTP(w, r)

	assert.True(t, proxy.Called, "허용된 요청은 origin으로 포워딩되어야 한다")
	assert.False(t, responder.Called, "허용된 요청에 대해 429를 반환하면 안 된다")
}

func TestServeHTTP_MatchingRequest_Denied(t *testing.T) {
	proxy := &MockProxy{}
	responder := &MockResponder{}
	handler := NewRateLimitHandler(
		&MockLimiter{
			IsTargetResult:  &types.ApiMatchResult{IsMatch: true, Identifier: "test"},
			IsAllowedResult: types.RateLimitDecision{Allowed: false, Remaining: 0, RetryAfterSec: 30},
		},
		proxy,
		responder,
		settings.RateLimiterConfig{
			Target: "http://example.com",
		},
		nil,
		&MockIdentifier{ClientID: "127.0.0.1"},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/test", nil)
	handler.ServeHTTP(w, r)

	assert.False(t, proxy.Called, "거부된 요청은 origin으로 포워딩되면 안 된다")
	assert.True(t, responder.Called, "거부된 요청에 대해 429를 반환해야 한다")
	assert.Equal(t, 0, responder.Remaining)
	assert.Equal(t, 30, responder.RetryAfter)
}

func TestServeHTTP_ClientLimitExceeded(t *testing.T) {
	mr := miniredis.RunT(t)
	rc := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	counterStore := storeredis.NewCounterStore(rc)

	proxy := &MockProxy{}
	responder := &MockResponder{}
	handler := NewRateLimitHandler(
		&MockLimiter{
			IsTargetResult:  &types.ApiMatchResult{IsMatch: true, Identifier: "test"},
			IsAllowedResult: types.RateLimitDecision{Allowed: true, Remaining: 10},
		},
		proxy,
		responder,
		settings.RateLimiterConfig{
			Target: "http://example.com",
			Client: settings.ClientLimit{Limit: 2, WindowSeconds: 60},
		},
		counterStore,
		&MockIdentifier{ClientID: "127.0.0.1"},
	)

	r := httptest.NewRequest("POST", "/api/test", nil)

	// 처음 2개 요청은 통과해야 한다
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		proxy.Called = false
		handler.ServeHTTP(w, r)
		assert.True(t, proxy.Called, "요청 %d는 포워딩되어야 한다", i+1)
	}

	// 3번째 요청은 글로벌 클라이언트 제한에 의해 차단되어야 한다
	w := httptest.NewRecorder()
	proxy.Called = false
	responder.Called = false
	handler.ServeHTTP(w, r)

	assert.False(t, proxy.Called, "클라이언트 제한 초과 시 포워딩되면 안 된다")
	assert.True(t, responder.Called, "클라이언트 제한 초과 시 429를 반환해야 한다")
}
