package limiter

import (
	"errors"
	"gate-limiter/internal/limiter/util"
	"gate-limiter/pkg/redisclient"
	"net/http"
	"strconv"
	"time"
)

// ==================================================
// MockRedisClient
// ==================================================
type MockRedisClient struct {
	size int
}

var _ redisclient.RedisClient = (*MockRedisClient)(nil)

func (m *MockRedisClient) RemoveOldEntries(key string, cutoff time.Time) error {
	return nil
}

func (m *MockRedisClient) AddToSortedSet(key, member string, score time.Time) error {
	return nil
}

func (m *MockRedisClient) GetZSetSize(key string) int {
	return m.size
}

func (m *MockRedisClient) GetOldestEntry(key string) (redisclient.Z, error) {
	return redisclient.Z{}, errors.New("not implemented")
}

func (m *MockRedisClient) RemoveOldEntry(key string, before time.Time) error {
	return nil
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{size: 3}
}

// ==================================================
// MockKeyGenerator
// ==================================================
type MockKeyGenerator struct{}

var _ util.KeyGenerator = (*MockKeyGenerator)(nil)

func (m *MockKeyGenerator) Make(identifier string, category string) string {
	return identifier + ":" + category
}

func NewMockKeyGenerator() *MockKeyGenerator {
	return &MockKeyGenerator{}
}

// ==================================================
// MockHttpLimitFailureResponder
// ==================================================
type MockHttpLimitFailureResponder struct {
	CalcRetryAfter func(key string) int
	RedisClient    redisclient.RedisClient
	KeyGenerator   util.KeyGenerator
}

var _ LimitResponder = (*MockHttpLimitFailureResponder)(nil)

func NewMockHttpLimitResponder() *MockHttpLimitFailureResponder {
	return &MockHttpLimitFailureResponder{
		CalcRetryAfter: func(key string) int {
			return 30 // 30second left to retry
		},
		RedisClient:  NewMockRedisClient(),
		KeyGenerator: NewMockKeyGenerator(),
	}
}

func (m *MockHttpLimitFailureResponder) RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int) {
	retryAfter := m.CalcRetryAfter(r.URL.Path) // return 30second

	w.Header().Set(XRateLimitRemaining, strconv.Itoa(remaining))
	w.Header().Set(XRateLimitReset, strconv.Itoa(5))
	w.Header().Set(XRateLimitRetryAfter, strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests) // HTTP 429 (too many requests)
}
