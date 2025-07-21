package limiter

import (
	"gate-limiter/pkg/redisclient"
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
	RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int)
}

type HttpLimitResponder struct {
	CalcRetryAfter func(key string) int
	RedisClient    redisclient.RedisClient
}

func NewHttpLimitResponder(redisClient redisclient.RedisClient) *HttpLimitResponder {
	h := &HttpLimitResponder{}
	h.CalcRetryAfter = h.defaultCalculateRetryAfter
	h.RedisClient = redisClient
	return h
}

func (h *HttpLimitResponder) RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int) {
	ipAddress := r.Header.Get(XForwardedFor)
	key := MakeRateLimitKey(ipAddress, "comment") // 이부분도 언젠가는 yml로 받아서 처리하길
	retryAfter := h.CalcRetryAfter(key)

	w.Header().Set(XRateLimitRemaining, strconv.Itoa(remaining))
	w.Header().Set(XRateLimitReset, strconv.Itoa(AllowedCount))
	w.Header().Set(XRateLimitRetryAfter, strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests) // HTTP 429 (too many requests)
}

func (h *HttpLimitResponder) defaultCalculateRetryAfter(key string) int {
	oldest, err := h.RedisClient.GetOldestEntry(key)
	if err != nil {
		log.Printf("fail to get oldest entry on key=[%s]\n", key)
	}

	oldestTime := time.Unix(int64(oldest.Score), 0)
	retryAt := oldestTime.Add(time.Minute * 1)
	now := time.Now()

	wait := retryAt.Sub(now).Seconds()
	if wait < 0 {
		return 0
	}
	return int(wait)
}
