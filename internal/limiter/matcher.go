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
	IsAllowed(ip string) (bool, int)
}

type HttpRateLimitMatcher struct{} // HTTP IP와 url Path를 기반으로 처리율 제한을 검사하는 구현체

func (rc *HttpRateLimitMatcher) IsTarget(method, url string) bool {
	return checkRateLimitTarget(method, url)
}

func (rc *HttpRateLimitMatcher) IsAllowed(ip string) (bool, int) {
	return isRequestAllowed(ip)
}

func isRequestAllowed(address string) (bool, int) {
	fmt.Printf("ip_adrress: [%s]를 검사합니다.\n", address)
	key := MakeRateLimitKey(address, "comment")

	var err error
	now := time.Now()

	err = redisclient.RemoveOldEntries(key, now.Add(-1*time.Minute))
	if err != nil {
		fmt.Println("error while removing old entries:", err)
	}
	err = redisclient.AddToSortedSet(key, now.String(), now)
	if err != nil {
		fmt.Println("error while adding to sorted set:", err)
	}
	size := redisclient.GetZSetSize(key)
	if size > AllowedCount {
		return false, 0
	}

	return true, AllowedCount - size
}

// checkRateLimitTarget 요청 대상 필터링 로직
// HTTP method와 url path를 보고 접근 제한 관리 대상인지 파악하는 메서드
// 추후 댓글작성 API 뿐만이 아니라 다른 API도 대응가능해야함.
func checkRateLimitTarget(method string, path string) bool {
	log.Printf("[%s] url_path:[%s] 를 검사합니다.", method, path)
	if method == http.MethodPost && commentPathPattern.MatchString(path) {
		return true
	}
	return false
}
