package limiter

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestHttpRateLimitMatcher_IsTarget(t *testing.T) {
	keyGenerator := &IpKeyGenerator{}
	httpRateLimitMatcher := NewHttpRateLimitMatcher(keyGenerator)

	passUrlPath := "/api/item/1/comment"
	result := httpRateLimitMatcher.IsTarget(http.MethodPost, passUrlPath)
	assert.True(t, result)

	refuseUrlPath := "/api/something"
	result = httpRateLimitMatcher.IsTarget(http.MethodPost, refuseUrlPath)
	assert.False(t, result)
}

func TestHttpRateLimitMatcher_IsAllowed(t *testing.T) {
	keyGenerator :=
}
