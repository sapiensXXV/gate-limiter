package limiter

import (
	"fmt"
	"gate-limiter/pkg/redisclient"
	"log"
	"net/http"
	"strconv"
	"time"
)

const x_forwarded_for = "X-Forwarded-For"
const ALLOWED_COUNT = 5

// 사용자로부터 받아온 요청을 처리하는 메서드
func HandleRateLimit(w http.ResponseWriter, r *http.Request) {
	log.Println()

}

func writeRateLimitExceededResponse(
	w http.ResponseWriter,
	r *http.Request,
	remainingCount int,
) {
	ipAdrress := r.Header.Get(x_forwarded_for)
	key := MakeRateLimitKey(ipAdrress, "comment") // 이부분도 언젠가는 yml로 받아서 처리하길
	retryAfter := redisclient.CalculateRetryAfter(key)

	w.Header().Set("X-Ratelimit-Remaining", strconv.Itoa(remainingCount))
	w.Header().Set("X-Ratelimit-Limit", strconv.Itoa(ALLOWED_COUNT))
	w.Header().Set("X-Ratelimit-Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
}

func IsRequestAllowed(address string) (bool, int) {
	fmt.Printf("ip_adrress: [%s]를 검사합니다.\n", address)
	key := MakeRateLimitKey(address, "comment")

	// SortedSet에 이미 값이 있는지 확인한다. 값이 있다면 해당 Set에 타임스탬프를 삽입한다.
	// 값이 없다면 Set을 만들어 타임스탬프를 삽입한다.
	var err error
	now := time.Now()

	// 오래된 요청 제거 (1분 이상 지난 요청)
	err = redisclient.RemoveOldEntries(key, now.Add(-1*time.Minute))
	if err != nil {
		fmt.Println("error while removing old entries:", err)
	}

	// 타임스탬프를 삽입하는 과정(공통)
	err = redisclient.AddToSortedSet(key, now.String(), now)
	if err != nil {
		fmt.Println("error while adding to sorted set:", err)
	}

	// 타임스탬프의 사이즈가 허용치를 초과하는지 검사한다.
	// 먼저 기존에 가지고 있는 데이터 중, 시간을 초과한 것이 있는지 확인한다. 시간을 초과한 타임스탬프는 삭제한다.
	size := redisclient.GetZSetSize(key)
	if size > ALLOWED_COUNT {
		// 허용량 보다 현재 셋의 사이즈가 크다면 요청을 거부한다.
		return false, 0
	}

	return true, ALLOWED_COUNT - size
}
