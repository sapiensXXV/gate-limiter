package limiter

type KeyGenerator interface {
	Make(identifier string, category string) string
}

// IpKeyGenerator Generate a key based on an IPv4 address.
type IpKeyGenerator struct{}

var _ KeyGenerator = (*IpKeyGenerator)(nil)

func (k *IpKeyGenerator) Make(identifier string, category string) string {
	return identifier + ":" + category
}
