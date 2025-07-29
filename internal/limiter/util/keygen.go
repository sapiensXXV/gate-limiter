package util

import "gate-limiter/config/settings"

type KeyGenerator interface {
	Make(address string, identifier string) string
}

// IpKeyGenerator Generate a key based on an IPv4 address.
type IpKeyGenerator struct {
	config settings.RateLimiterConfig
}

var _ KeyGenerator = (*IpKeyGenerator)(nil)

func NewIpKeyGenerator() *IpKeyGenerator {
	return &IpKeyGenerator{}
}

func (k *IpKeyGenerator) Make(address string, category string) string {
	return k.config.Strategy + ":" + address + ":" + category
}
