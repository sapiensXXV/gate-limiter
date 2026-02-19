package limiter

import (
	"gate-limiter/internal/limiter/types"
	"net/http"
)

// MockLimiter implements types.RateLimiter for testing.
type MockLimiter struct {
	IsTargetResult  *types.ApiMatchResult
	IsAllowedResult types.RateLimitDecision
}

var _ types.RateLimiter = (*MockLimiter)(nil)

func (m *MockLimiter) IsTarget(method, requestPath string) *types.ApiMatchResult {
	return m.IsTargetResult
}

func (m *MockLimiter) IsAllowed(ip string, api *types.ApiMatchResult, qr *types.QueuedRequest) types.RateLimitDecision {
	return m.IsAllowedResult
}

// MockProxy implements types.ProxyHandler for testing.
type MockProxy struct {
	Called bool
}

var _ types.ProxyHandler = (*MockProxy)(nil)

func (m *MockProxy) ToOrigin(w http.ResponseWriter, r *http.Request, origin string) {
	m.Called = true
	w.WriteHeader(http.StatusOK)
}

// MockIdentifier implements ClientIdentifier for testing.
type MockIdentifier struct {
	ClientID string
}

var _ ClientIdentifier = (*MockIdentifier)(nil)

func (m *MockIdentifier) Identify(_ http.ResponseWriter, _ *http.Request) string {
	return m.ClientID
}

// MockResponder implements LimitResponder for testing.
type MockResponder struct {
	Called     bool
	Remaining  int
	RetryAfter int
}

var _ LimitResponder = (*MockResponder)(nil)

func (m *MockResponder) RespondRateLimitExceeded(w http.ResponseWriter, r *http.Request, remaining int, retryAfter int) {
	m.Called = true
	m.Remaining = remaining
	m.RetryAfter = retryAfter
	w.WriteHeader(http.StatusTooManyRequests)
}
