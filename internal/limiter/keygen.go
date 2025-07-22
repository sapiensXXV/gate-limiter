package limiter

type KeyGenerator interface {
	Make(identifier string, category string) string
}

type IpKeyGenerator struct{}

var _ KeyGenerator = (*IpKeyGenerator)(nil)

func (k *IpKeyGenerator) Make(identifier string, category string) string {
	return identifier + ":" + category
}
