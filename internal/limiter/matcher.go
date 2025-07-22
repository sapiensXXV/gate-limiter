package limiter

import (
	"fmt"
	"gate-limiter/pkg/redisclient"
	"log"
	"net/http"
	"regexp"
	"time"
)

var commentPathPattern = regexp.MustCompile(`^/api/item/[\w-]+/comment$`)

type RateLimitMatcher interface {
	IsTarget(method, url string) bool
	IsAllowed(ip string) (bool, int) // (허용 여부, 남은 허용 요청량)
}

type HttpRateLimitMatcher struct {
	KeyGenerator KeyGenerator
	RedisClient  redisclient.RedisClient
} // HTTP IP와 url Path를 기반으로 처리율 제한을 검사하는 구현체

var _ RateLimitMatcher = (*HttpRateLimitMatcher)(nil)

func NewHttpRateLimitMatcher(keyGenerator KeyGenerator) *HttpRateLimitMatcher {
	h := &HttpRateLimitMatcher{}
	h.KeyGenerator = keyGenerator
	return h
}

func (rc *HttpRateLimitMatcher) IsTarget(method, urlPath string) bool {
	log.Printf("[%s] url_path:[%s] 를 검사합니다.", method, urlPath)
	if method == http.MethodPost && commentPathPattern.MatchString(urlPath) {
		return true
	}
	return false
}

func (rc *HttpRateLimitMatcher) IsAllowed(ip string) (bool, int) {
	fmt.Printf("ip_adrress: [%s]를 검사합니다.\n", ip)
	key := rc.KeyGenerator.Make(ip, "comment")

	var err error
	now := time.Now()

	err = rc.RedisClient.RemoveOldEntries(key, now.Add(-1*time.Minute))
	if err != nil {
		log.Println("error while removing old entries:", err)
	}
	err = rc.RedisClient.AddToSortedSet(key, now.String(), now)
	if err != nil {
		log.Println("error while adding to sorted set:", err)
	}
	size := rc.RedisClient.GetZSetSize(key)
	if size > AllowedCount {
		return false, 0
	}

	return true, AllowedCount - size
}
