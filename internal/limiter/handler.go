package limiter

import (
	"gate-limiter/config/settings"
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
		return
	}
	result := h.Limiter.IsTarget(r.Method, r.URL.Path)
	policy := h.Config.Strategy

	if !result.IsMatch {
		h.Proxy.ToOrigin(w, r, h.Config.Target)
		return
	}

	clientID := r.Header.Get(h.Config.Identity.Header)

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
