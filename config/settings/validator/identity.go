package validator

import "fmt"

var identities = map[string]struct{}{
	"ipv4":   {},
	"cookie": {},
}

func ValidateIdentity(key string, header string) error {
	// 키 검사부
	if key == "" {
		return fmt.Errorf("rateLimiter.identity.key 가 설정되어 있지 않습니다. 유효한 값: ipv4, cookie")
	}
	if _, ok := identities[key]; !ok {
		return fmt.Errorf("알 수 없는 rateLimiter.identity.key 입니다. 현재 값: %s, 유효한 값: ipv4, cookie", key)
	}

	// 헤더 검사부 (cookie 방식에서는 header 불필요)
	if key == "ipv4" && header == "" {
		return fmt.Errorf("rateLimiter.identity.header 가 설정되어 있지 않습니다 (identity.key=ipv4 일 때 필수)")
	}

	return nil
}
