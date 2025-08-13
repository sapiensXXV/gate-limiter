package limiter

import (
	"gate-limiter/internal/limiter/types"
	"log"
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
	log.Printf("원래 요청 경로: [%s %s%s] 로 요청을 재전달합니다\n", r.Method, target, r.URL.RequestURI())

	proxy.ServeHTTP(w, r)
}
