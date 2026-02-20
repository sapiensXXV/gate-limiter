package limiter

import (
	"encoding/json"
	"fmt"
	config_ratelimiter "gate-limiter/config/settings"
	"gate-limiter/internal/middleware"
	"net/http"
	"strconv"
	"strings"
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
	r *http.Request,
	remaining int,
	retryAfter int,
) {
	resetAt := time.Now().Add(time.Duration(retryAfter) * time.Second).Unix()

	w.Header().Set(XRateLimitRemaining, strconv.Itoa(remaining))
	w.Header().Set(XRateLimitReset, strconv.FormatInt(resetAt, 10))
	w.Header().Set(XRateLimitRetryAfter, strconv.Itoa(retryAfter))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)

	requestID := middleware.RequestIDFrom(r.Context())
	message := rateLimitMessage(r, retryAfter)

	responseBody := map[string]interface{}{
		"error":               "Too Many Requests",
		"message":             message,
		"retry_after_seconds": retryAfter,
		"remaining":           remaining,
		"request_id":          requestID,
	}

	json.NewEncoder(w).Encode(responseBody)
}

// rateLimitMessage returns a localized rate limit message based on Accept-Language header.
func rateLimitMessage(r *http.Request, retryAfter int) string {
	if prefersKorean(r.Header.Get("Accept-Language")) {
		return fmt.Sprintf("요청 한도를 초과했습니다. %d초 후에 다시 시도해주세요.", retryAfter)
	}
	return fmt.Sprintf("Rate limit exceeded. Please retry after %d seconds.", retryAfter)
}

// prefersKorean checks if the Accept-Language header indicates Korean preference.
func prefersKorean(acceptLang string) bool {
	if acceptLang == "" {
		return false
	}
	for _, part := range strings.Split(acceptLang, ",") {
		lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if strings.EqualFold(lang, "ko") || strings.HasPrefix(strings.ToLower(lang), "ko-") {
			return true
		}
	}
	return false
}
