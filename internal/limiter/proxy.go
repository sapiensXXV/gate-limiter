package limiter

import (
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type DefaultProxyHandler struct{}

var _ types.ProxyHandler = (*DefaultProxyHandler)(nil)

func NewDefaultProxyHandler() *DefaultProxyHandler {
	return &DefaultProxyHandler{}
}

func (dph *DefaultProxyHandler) ToOrigin(w http.ResponseWriter, r *http.Request, origin string) {
	logger := middleware.LoggerFrom(r.Context())

	target, err := url.Parse(origin)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Path = r.URL.Path
		req.URL.RawQuery = r.URL.RawQuery
		req.Header.Set("X-Forwarded-For", r.Header.Get(XForwardedFor))
	}
	logger.Info("proxying request", "target", target.String(), "uri", r.URL.RequestURI())

	proxy.ServeHTTP(w, r)
}
