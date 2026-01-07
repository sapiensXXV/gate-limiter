package limiter

import (
	"gate-limiter/internal/limiter/types"
	"net/http"
)

type MockLimiter struct {
	TargetStub  types.TargetResult
	AllowedStub types.AllowedResult
}

func (m *MockLimiter) IsTarget(method, url string) types.TargetResult {
	return m.TargetStub
}

func (m *MockLimiter) IsAllowed(id string, target types.TargetResult, queued *types.QueuedRequest) types.AllowedResult {
	return m.AllowedStub
}

type MockProxy struct {
	CalledToOrigin bool
}

func (m *MockProxy) ToOrigin(w http.ResponseWriter, r *http.Request, target string) types.ProxyHandler {
	m.CalledToOrigin = true
}

type MockResponder struct {
	CalledBlocked bool
}
