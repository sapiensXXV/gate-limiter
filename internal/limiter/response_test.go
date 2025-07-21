package limiter

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHttpLimitResponder_RespondRateLimitExceeded(t *testing.T) {
	responder := &HttpLimitResponder{}
	remaining := 0
	ip := "127.0.0.1"
	req := httptest.NewRequest(http.MethodPost, "/api/comment", nil)
	req.Header.Set(XForwardedFor, ip)
	rr := httptest.NewRecorder()

	//when
	responder.RespondRateLimitExceeded(rr, req, remaining)

	//then
	result := rr.Result()
	defer result.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, result.StatusCode)
	assert.Equal(t, strconv.Itoa(remaining), result.Header.Get(XRateLimitRemaining))
	assert.Equal(t, strconv.Itoa(AllowedCount), result.Header.Get(XRateLimitReset))
	assert.Equal(t, "60", result.Header.Get(XRateLimitRetryAfter))

}
