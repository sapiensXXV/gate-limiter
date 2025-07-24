package strategy

import (
	"gate-limiter/internal/limiter"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestHttpRateLimitMatcher_IsTarget(t *testing.T) {
	keyGenerator := &limiter.IpKeyGenerator{} // KeyGenerator
	rc := &limiter.MockRedisClient{}          // RedisClient
	// TODO config.yml mock 설정정보 객체
	httpRateLimitMatcher := NewSlidingWindowLogLimiter(keyGenerator, rc, nil)

	passUrlPath := "/api/item/1/comment"
	result := httpRateLimitMatcher.IsTarget(http.MethodPost, passUrlPath)
	assert.True(t, result)

	refuseUrlPath := "/api/something"
	result = httpRateLimitMatcher.IsTarget(http.MethodPost, refuseUrlPath)
	assert.False(t, result)
}

func TestHttpRateLimitMatcher_IsAllowed_Allowed(t *testing.T) {
	keyGenerator := &limiter.MockKeyGenerator{}
	rc := &limiter.MockRedisClient{size: limiter.AllowedCount - 2} // 허용치보다 2개 여유 있는 상태
	matcher := NewHttpRateLimitMatcher(keyGenerator, rc)

	allowed, remaining := matcher.IsAllowed("192.0.2.1")
	assert.True(t, allowed)
	assert.Equal(t, limiter.AllowedCount-3, remaining)
}

func TestHttpRateLimitMatcher_IsAllowed_Refused(t *testing.T) {
	keyGenerator := &limiter.MockKeyGenerator{}
	rc := &limiter.MockRedisClient{size: limiter.AllowedCount + 1} // 허용치보다 하나가 많은 상태
	matcher := NewHttpRateLimitMatcher(keyGenerator, rc)

	allowed, remaining := matcher.IsAllowed("192.0.2.1")
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining) // 현재 허용된 요청 갯수 0개 (요청불가)
}
