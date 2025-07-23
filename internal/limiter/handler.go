package limiter

import (
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"log"
	"net/http"
)

const XForwardedFor = "X-Forwarded-For"
const AllowedCount = 5

type RateLimitHandler struct {
	Matcher   RateLimitMatcher
	Proxy     ProxyHandler
	Responder LimitResponder
	Config    config_ratelimiter.RateLimiterConfig
}

var _ http.Handler = (*RateLimitHandler)(nil)

func NewRateLimitHandler(
	matcher RateLimitMatcher,
	proxy ProxyHandler,
	responder LimitResponder,
	config config_ratelimiter.RateLimiterConfig,
) *RateLimitHandler {
	return &RateLimitHandler{
		Matcher:   matcher,
		Proxy:     proxy,
		Responder: responder,
		Config:    config,
	}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.Matcher.IsTarget(r.Method, r.URL.String()) {
		log.Printf("[%s] url_path:[%s] 는 검사 대상이 아닙니다.", r.Method, r.URL.Path)
		h.Proxy.ToOrigin(w, r)
		return
	}

	log.Printf("[%s] url_path:[%s] 는 검사 대상입니다.", r.Method, r.URL.Path)
	allowed, remaining := h.Matcher.IsAllowed(r.Header.Get(h.Config.Identity.Header))
	if !allowed {
		log.Printf("[%s] url_path:[%s] 는 허용치를 초과하였습니다.", r.Method, r.URL.Path)
		h.Responder.RespondRateLimitExceeded(w, r, remaining)
		return
	}

	h.Proxy.ToOrigin(w, r)
}
