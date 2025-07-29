package limiter

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/internal/limiter/types"
	"net/http"
)

const XForwardedFor = "X-Forwarded-For"
const AllowedCount = 5

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
	isTarget, api := h.Limiter.IsTarget(r.Method, r.URL.String())

	if !isTarget {
		h.Proxy.ToOrigin(w, r, api.Target)
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
			ll.IsAllowed(r.Header.Get(h.Config.Identity.Header), api, queued)
		}
	} else {
		// token_bucket, sliding_window_log, sliding_window_counter
		// 다른 알고리즘의 경우에는 QueuedRequest를 사용하지 않는다.
		allowed, remaining := h.Limiter.IsAllowed(r.Header.Get(h.Config.Identity.Header), api, nil)
		if !allowed {
			h.Responder.RespondRateLimitExceeded(w, r, remaining)
			return
		}
	}

	h.Proxy.ToOrigin(w, r, api.Target)
}
