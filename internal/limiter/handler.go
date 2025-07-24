package limiter

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter/strategy"
	"log"
	"net/http"
)

const XForwardedFor = "X-Forwarded-For"
const AllowedCount = 5

type RateLimitHandler struct {
	Limiter   strategy.RateLimiter
	Proxy     ProxyHandler
	Responder LimitResponder
	Config    config_ratelimiter.RateLimiterConfig
}

var _ http.Handler = (*RateLimitHandler)(nil)

func NewRateLimitHandler(
	matcher strategy.RateLimiter,
	proxy ProxyHandler,
	responder LimitResponder,
	config config_ratelimiter.RateLimiterConfig,
) *RateLimitHandler {
	return &RateLimitHandler{
		Limiter:   matcher,
		Proxy:     proxy,
		Responder: responder,
		Config:    config,
	}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isTarget, api := h.Limiter.IsTarget(r.Method, r.URL.String())

	if !isTarget {
		//log.Printf("[%s] url_path:[%s] 는 검사 대상이 아닙니다.", r.Method, r.URL.Path)
		h.Proxy.ToOrigin(w, r, api.Target)
		return
	}

	//log.Printf("[%s] url_path:[%s] 는 검사 대상입니다.", r.Method, r.URL.Path)
	allowed, remaining := h.Limiter.IsAllowed(r.Header.Get(h.Config.Identity.Header), api)
	if !allowed {
		log.Printf("[%s] url_path:[%s] 는 허용치를 초과하였습니다.", r.Method, r.URL.Path)
		h.Responder.RespondRateLimitExceeded(w, r, remaining)
		return
	}

	h.Proxy.ToOrigin(w, r, api.Target)
}
