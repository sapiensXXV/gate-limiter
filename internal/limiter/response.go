package limiter

import (
	config_ratelimiter "gate-limiter/config/settings"
	"gate-limiter/internal/limiter/util"
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
	KeyGenerator   util.KeyGenerator
	Config         config_ratelimiter.RateLimiterConfig
}

func NewHttpLimitResponder(
	calcRetryAfter func(key string) int,
	redisClient redisclient.RedisClient,
	keyGenerator util.KeyGenerator,
	config config_ratelimiter.RateLimiterConfig,
) *HttpLimitResponder {
	h := &HttpLimitResponder{}
	if calcRetryAfter != nil {
		h.CalcRetryAfter = calcRetryAfter
	} else {
		h.CalcRetryAfter = h.defaultCalculateRetryAfter
	}
	h.RedisClient = redisClient
	h.KeyGenerator = keyGenerator
	h.Config = config
	return h
}

func (h *HttpLimitResponder) RespondRateLimitExceeded(
	w http.ResponseWriter,
	r *http.Request,
	remaining int,
) {
	ipAddress := r.Header.Get(h.Config.Identity.Header)
	// TODO Apis에 수동으로 인덱스를 주어야하는가? 설정한 목록대로 다 처리하려면 결국 반복문을 돌면서 ResponseRateLimitExceeded 메서드를 호출해야하는데
	// 매개변수로 key, allow_count, header를 다 받아야하나? 그건 좀 에반데.
	key := h.KeyGenerator.Make(ipAddress, h.Config.Apis[0].Identifier)
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
