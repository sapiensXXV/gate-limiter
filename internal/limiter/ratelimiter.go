package limiter

import (
	"gate-limiter/pkg/redisclient"
	"log"
	"time"
)

// CalculateRetryAfter 함수는 주어진 Redis Sorted Set 키를 기반으로,
// 가장 오래된 요청의 타임스탬프를 조회하여 다음 요청까지 기다려야 하는 시간을 초 단위로 계산합니다.
//
// 이 함수는 레이트 리밋 윈도우(1분) 안에서 가장 오래된 요청을 기준으로
// 얼마나 기다려야 윈도우가 갱신되는지를 판단합니다.
//
// 매개변수:
//   - key: Redis Sorted Set의 키 (예: "192.168.0.1:comment")
//
// 반환값:
//   - int: 재요청까지 기다려야 하는 시간(초).
//     즉시 요청 가능한 경우 0, 오류가 발생했거나 키가 없는 경우 60초를 반환합니다.
func CalculateRetryAfter(key string) int {
	vals, err := redisclient.GetZRangeWithScores(key, 0, 0)
	if err != nil || len(vals) == 0 {
		log.Println("error fetching oldest entry:", err)
		return 60
	}
	oldest := vals[0].Score
	oldestTime := time.Unix(int64(oldest), 0)
	retryAt := oldestTime.Add(time.Minute * 1)
	now := time.Now()

	wait := retryAt.Sub(now).Seconds()
	if wait < 0 {
		return 0
	}
	return int(wait)
}
