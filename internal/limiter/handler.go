package limiter

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/metrics"
	"log"
	"net/http"
)

const XForwardedFor = "X-Forwarded-For"

type RateLimitHandler struct {
	Limiter   types.RateLimiter
	Proxy     types.ProxyHandler
	Responder LimitResponder
	Config    settings.RateLimiterConfig
}

var _ http.Handler = (*RateLimitHandler)(nil)

func NewRateLimitHandler(
	limiter types.RateLimiter,
	proxy types.ProxyHandler,
	responder LimitResponder,
	config settings.RateLimiterConfig,
) *RateLimitHandler {
	return &RateLimitHandler{
		Limiter:   limiter,
		Proxy:     proxy,
		Responder: responder,
		Config:    config,
	}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Limiter == nil {
		log.Println("RateLimitHandler.Limiter is nil!")
	}
	result := h.Limiter.IsTarget(r.Method, r.URL.String())
	policy := h.Config.Strategy

	if !result.IsMatch {
		h.Proxy.ToOrigin(w, r, h.Config.Target)
		return
	}

	var queued *types.QueuedRequest
	if h.Config.Strategy == "leaky_bucket" {
		// leaky_bucket 알고리즘을 사용하는 경우 현재 요청/응답 정보를 큐에 넘겨야한다.
		queued = &types.QueuedRequest{
			Writer:  w,
			Request: r,
		}
		if ll, ok := h.Limiter.(*strategy.LeakyBucketLimiter); ok {
			ll.IsAllowed(r.Header.Get(h.Config.Identity.Header), result, queued)
		}
	} else {
		// token_bucket, sliding_window_log, sliding_window_counter
		// 다른 알고리즘의 경우에는 QueuedRequest를 사용하지 않는다.
		decision := h.Limiter.IsAllowed(r.Header.Get(h.Config.Identity.Header), result, nil)
		if !decision.Allowed {
			h.Responder.RespondRateLimitExceeded(w, r, decision.Remaining, decision.RetryAfterSec)
			metrics.ObserveBlocked(policy) // 메트릭 집계 -> 블록
			return
		}
		metrics.ObserveAllowed(policy) // 메트릭 집계 -> 통과
	}

	h.Proxy.ToOrigin(w, r, h.Config.Target)
}
