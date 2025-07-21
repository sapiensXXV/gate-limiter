package limiter

import (
	"net/http"
	"strconv"
)

type LimitResponder interface {
	RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int)
}

type HttpLimitResponder struct{}

func (h *HttpLimitResponder) RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int) {
	ipAddress := r.Header.Get(XForwardedFor)
	key := MakeRateLimitKey(ipAddress, "comment") // 이부분도 언젠가는 yml로 받아서 처리하길
	retryAfter := CalculateRetryAfter(key)

	w.Header().Set("X-Ratelimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-Ratelimit-Limit", strconv.Itoa(AllowedCount))
	w.Header().Set("X-Ratelimit-Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests) // HTTP 429 (too many requests)
}
