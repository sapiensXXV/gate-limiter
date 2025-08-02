package validator

import "fmt"

var identities = map[string]struct{}{
	"ipv4": {},
}

func ValidateIdentity(key string, header string) error {
	// 키 검사부
	if key == "" {
		return fmt.Errorf("rateLimiter.identity.key 가 설정되어 있지 않습니다. 유효한 값: %s\n", "ipv4")
	} else if key != "ipv4" {
		return fmt.Errorf("알 수 없는 rateLimiter.identity.key 입니다. 현재 값: %s, 유효한 값: %s\n", key, "ipv4")
	}

	// 헤더 검사부
	if header == "" {
		return fmt.Errorf("rateLimiter.identity.header 가 설정되어 있지 않습니다. 유효한 값: %s\n", "GET, POST, PUT, DELETE")
	}

	return nil
}
