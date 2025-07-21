package limiter

import "fmt"

// Redis 키 생성 유틸 함수
func MakeRateLimitKey(ip string, category string) string {
	return fmt.Sprintf("%s_%s", ip, category)
}
