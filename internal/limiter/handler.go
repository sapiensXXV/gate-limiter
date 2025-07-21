package limiter

import (
	"net/http"
)

const XForwardedFor = "X-Forwarded-For"
const AllowedCount = 5

type RateLimitHandler struct {
	Matcher   RateLimitMatcher
	Proxy     ProxyHandler
	Responder LimitResponder
}

var _ http.Handler = (*RateLimitHandler)(nil)

func NewRateLimitHandler(
	matcher RateLimitMatcher,
	proxy ProxyHandler,
	responder LimitResponder,
) *RateLimitHandler {
	return &RateLimitHandler{
		Matcher:   matcher,
		Proxy:     proxy,
		Responder: responder,
	}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Matcher.IsTarget(r.Method, r.URL.String()) {
		h.Proxy.ToOrigin(w, r)
		return
	}

	allowed, remaining := h.Matcher.IsAllowed(r.Header.Get(XForwardedFor))
	if !allowed {
		h.Responder.RespondRateLimitExceeded(w, r, remaining)
		return
	}

	h.Proxy.ToOrigin(w, r)
}
