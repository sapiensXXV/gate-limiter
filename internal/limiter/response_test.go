package limiter

import (
	"encoding/json"
	"gate-limiter/config/settings"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpLimitResponder_RespondRateLimitExceeded(t *testing.T) {
	responder := NewHttpLimitResponder(nil, nil, settings.RateLimiterConfig{})

	writer := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/test", nil)

	remaining := 0
	retryAfter := 30
	responder.RespondRateLimitExceeded(writer, request, remaining, retryAfter)

	// HTTP 상태코드 검증
	assert.Equal(t, http.StatusTooManyRequests, writer.Code)

	// Rate Limit 헤더 검증
	assert.Equal(t, strconv.Itoa(remaining), writer.Header().Get(XRateLimitRemaining))
	assert.Equal(t, strconv.Itoa(retryAfter), writer.Header().Get(XRateLimitRetryAfter))
	assert.NotEmpty(t, writer.Header().Get(XRateLimitReset))

	// Content-Type 검증
	assert.Equal(t, "application/json; charset=utf-8", writer.Header().Get("Content-Type"))

	// JSON 응답 본문 검증
	var body map[string]interface{}
	err := json.NewDecoder(writer.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "Too Many Requests", body["error"])
	assert.Equal(t, float64(retryAfter), body["retry_after"])
}

func TestHttpLimitResponder_RemainingHeader(t *testing.T) {
	responder := NewHttpLimitResponder(nil, nil, settings.RateLimiterConfig{})

	writer := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/test", nil)

	remaining := 5
	retryAfter := 10
	responder.RespondRateLimitExceeded(writer, request, remaining, retryAfter)

	assert.Equal(t, strconv.Itoa(remaining), writer.Header().Get(XRateLimitRemaining))
	assert.Equal(t, strconv.Itoa(retryAfter), writer.Header().Get(XRateLimitRetryAfter))
}
