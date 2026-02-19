package limiter

import (
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/metrics"
	"log"
	"math"
	"net/http"
	"time"
)

const XForwardedFor = "X-Forwarded-For"

// clientLimitLuaScript 는 글로벌 클라이언트 제한을 위한 Lua 스크립트이다.
// Fixed Window Counter 방식으로 클라이언트의 전체 요청 수를 원자적으로 카운트한다.
const clientLimitLuaScript = `
local current = redis.call("INCR", KEYS[1])
if tonumber(current) == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`

type RateLimitHandler struct {
	Limiter     types.RateLimiter
	Proxy       types.ProxyHandler
	Responder   LimitResponder
	Config      settings.RateLimiterConfig
	RedisClient types.RedisClient
}

var _ http.Handler = (*RateLimitHandler)(nil)

func NewRateLimitHandler(
	limiter types.RateLimiter,
	proxy types.ProxyHandler,
	responder LimitResponder,
	config settings.RateLimiterConfig,
	redisClient types.RedisClient,
) *RateLimitHandler {
	return &RateLimitHandler{
		Limiter:     limiter,
		Proxy:       proxy,
		Responder:   responder,
		Config:      config,
		RedisClient: redisClient,
	}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Limiter == nil {
		log.Println("RateLimitHandler.Limiter is nil!")
		return
	}

	policy := h.Config.Strategy
	clientID := r.Header.Get(h.Config.Identity.Header)

	// 글로벌 클라이언트 제한 검사
	if exceeded, _, retryAfter := h.isClientLimitExceeded(clientID); exceeded {
		h.Responder.RespondRateLimitExceeded(w, r, 0, retryAfter)
		metrics.ObserveBlocked(policy, "client_limit_exceeded")
		return
	}

	result := h.Limiter.IsTarget(r.Method, r.URL.Path)

	if !result.IsMatch {
		h.Proxy.ToOrigin(w, r, h.Config.Target)
		return
	}

	// leaky_bucket은 내부적으로 queueing/wait를 하기 때문에 Request Context를 넘겨준다
	var queued *types.QueuedRequest
	if h.Config.Strategy == "leaky_bucket" {
		queued = &types.QueuedRequest{
			Request: r,
		}
	}

	decision := h.Limiter.IsAllowed(clientID, result, queued)
	if !decision.Allowed {
		h.Responder.RespondRateLimitExceeded(w, r, decision.Remaining, decision.RetryAfterSec)
		metrics.ObserveBlocked(policy, "허용치 초과")
		return
	}
	metrics.ObserveAllowed(policy)

	h.Proxy.ToOrigin(w, r, h.Config.Target)
}

// isClientLimitExceeded 글로벌 클라이언트 제한을 검사한다.
// Config.Client.Limit 이 설정되어 있지 않으면(0 이하) 검사를 건너뛴다.
func (h *RateLimitHandler) isClientLimitExceeded(clientID string) (exceeded bool, remaining int, retryAfterSec int) {
	if h.Config.Client.Limit <= 0 || h.Config.Client.WindowSeconds <= 0 || h.RedisClient == nil {
		return false, 0, 0
	}

	windowDuration := time.Duration(h.Config.Client.WindowSeconds) * time.Second
	windowStart := time.Now().Truncate(windowDuration)
	key := fmt.Sprintf("client:%s:%d", clientID, windowStart.Unix())

	result, err := h.RedisClient.Eval(clientLimitLuaScript, []string{key}, h.Config.Client.WindowSeconds)
	if err != nil {
		log.Printf("client limit check error: %v", err)
		return false, 0, 0 // fail-open: Redis 장애 시 요청을 차단하지 않음
	}

	cnt, ok := result.(int64)
	if !ok {
		log.Printf("client limit: unexpected result type: %T", result)
		return false, 0, 0
	}

	if cnt > int64(h.Config.Client.Limit) {
		retryAt := windowStart.Add(windowDuration)
		wait := time.Until(retryAt)
		sec := int(math.Ceil(wait.Seconds()))
		if sec < 0 {
			sec = 0
		}
		return true, 0, sec
	}

	return false, h.Config.Client.Limit - int(cnt), 0
}
