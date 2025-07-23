package limiter

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHttpLimitResponder_RespondRateLimitExceeded(t *testing.T) {
	// 전달해준 내용으로 응답객체를 잘 써주는지 확인한다.
	calcRetryAfter := func(key string) int {
		return 30
	}
	redisClient := NewMockRedisClient()
	keyGenerator := NewMockKeyGenerator()
	//TODO config 설정
	responder := NewHttpLimitResponder(calcRetryAfter, redisClient, keyGenerator, nil)

	writer := httptest.NewRecorder()
	request := &http.Request{Header: http.Header{}}
	request.Header.Set(XForwardedFor, "192.0.1.0")

	remaining := 0
	responder.RespondRateLimitExceeded(writer, request, remaining)

	assert.Equal(t, writer.Header().Get(XRateLimitRemaining), strconv.Itoa(remaining))
	assert.Equal(t, writer.Header().Get(XRateLimitReset), strconv.Itoa(AllowedCount))
	assert.Equal(t, writer.Header().Get(XRateLimitRetryAfter), strconv.Itoa(30))
	assert.Equal(t, writer.Code, http.StatusTooManyRequests)

}
