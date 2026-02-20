package validator

import "fmt"

var identities = map[string]struct{}{
	"ipv4":   {},
	"cookie": {},
}

func ValidateIdentity(key string, header string) error {
	if key == "" {
		return fmt.Errorf("rateLimiter.identity.key: not configured. allowed: ipv4, cookie")
	}
	if _, ok := identities[key]; !ok {
		return fmt.Errorf("rateLimiter.identity.key: unknown value %q. allowed: ipv4, cookie", key)
	}

	if key == "ipv4" && header == "" {
		return fmt.Errorf("rateLimiter.identity.header: required when identity.key is \"ipv4\"")
	}

	return nil
}
