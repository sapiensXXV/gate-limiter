package limiter

import (
	"encoding/json"
	config_ratelimiter "gate-limiter/config/settings"
	"net/http"
	"strconv"
	"time"
)

const (
	XRateLimitRemaining  = "X-RateLimit-Remaining"
	XRateLimitReset      = "X-RateLimit-Reset"
	XRateLimitRetryAfter = "X-RateLimit-Retry-After"
)

type LimitResponder interface {
	RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int, retryAfter int)
}

type HttpLimitResponder struct {
	Config config_ratelimiter.RateLimiterConfig
}

func NewHttpLimitResponder(config config_ratelimiter.RateLimiterConfig) *HttpLimitResponder {
	return &HttpLimitResponder{Config: config}
}

func (h *HttpLimitResponder) RespondRateLimitExceeded(
	w http.ResponseWriter,
	_ *http.Request,
	remaining int,
	retryAfter int,
) {
	resetAt := time.Now().Add(time.Duration(retryAfter) * time.Second).Unix()

	w.Header().Set(XRateLimitRemaining, strconv.Itoa(remaining))
	w.Header().Set(XRateLimitReset, strconv.FormatInt(resetAt, 10))
	w.Header().Set(XRateLimitRetryAfter, strconv.Itoa(retryAfter))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)

	responseBody := map[string]interface{}{
		"error":       "Too Many Requests",
		"message":     "요청 한도를 초과했습니다. 잠시 후 다시 시도해주세요",
		"retry_after": retryAfter,
	}

	json.NewEncoder(w).Encode(responseBody)
}
