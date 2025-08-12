package limiter

import (
	config_ratelimiter "gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
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
	RedisClient  types.RedisClient
	KeyGenerator util.KeyGenerator
	Config       config_ratelimiter.RateLimiterConfig
}

func NewHttpLimitResponder(
	redisClient types.RedisClient,
	keyGenerator util.KeyGenerator,
	config config_ratelimiter.RateLimiterConfig,
) *HttpLimitResponder {
	h := &HttpLimitResponder{}
	h.RedisClient = redisClient
	h.KeyGenerator = keyGenerator
	h.Config = config
	return h
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
	w.WriteHeader(http.StatusTooManyRequests) // HTTP 429 (too many requests)
}
