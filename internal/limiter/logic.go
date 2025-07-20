package limiter

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

var commentPathPattern = regexp.MustCompile(`^/api/item/[\w-]+/comment$`)

// CheckRateLimitTarget
// HTTP method와 url path를 보고 접근 제한 관리 대상인지 파악하는 메서드
// 추후 댓글작성 API 뿐만이 아니라 다른 API도 대응가능해야함.
func CheckRateLimitTarget(method string, path string) bool {
	if method == http.MethodPost && commentPathPattern.MatchString(path) {
		return true
	}
	return false
}

func ProxyToOrigin(w http.ResponseWriter, r *http.Request) {
	target, err := url.Parse("http://localhost:8080")
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
		req.Header.Set("X-Forwarded-For", r.Header.Get(x_forwarded_for))
	}
	fmt.Printf("원래 요청 경로: [%s %s%s] 로 요청을 재전달합니다.®\n", r.Method, target, r.URL.RequestURI())

	proxy.ServeHTTP(w, r)
}

func MakeRateLimitKey(ip string, category string) string {
	return fmt.Sprintf("%s_%s", ip, category)
}
