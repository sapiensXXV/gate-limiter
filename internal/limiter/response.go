package limiter

import (
	config_ratelimiter "gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"
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
	r *http.Request,
	remaining int,
	retryAfter int,
) {
	ipAddress := r.Header.Get(h.Config.Identity.Header)
	// TODO Apis에 수동으로 인덱스를 주어야하는가? 설정한 목록대로 다 처리하려면 결국 반복문을 돌면서 ResponseRateLimitExceeded 메서드를 호출해야하는데
	// 매개변수로 key, allow_count, header를 다 받아야하나? 그건 좀 에반데.
	key := h.KeyGenerator.Make(ipAddress, h.Config.Apis[0].Identifier)
	w.Header().Set(XRateLimitRemaining, strconv.Itoa(remaining))
	w.Header().Set(XRateLimitReset, strconv.Itoa(AllowedCount))
	w.Header().Set(XRateLimitRetryAfter, strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests) // HTTP 429 (too many requests)
}
