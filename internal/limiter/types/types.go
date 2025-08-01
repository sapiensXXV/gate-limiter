package types

import "net/http"

type RateLimiter interface {
	IsTarget(method, requestPath string) *ApiMatchResult
	IsAllowed(ip string, api *ApiMatchResult, queuedRequest *QueuedRequest) RateLimitDecision
}

type RateLimitDecision struct {
	Allowed       bool // 허용 여부
	Remaining     int  // 남아있는 허용 용량(버킷, 윈도우, 맵 등)
	RetryAfterSec int  // 재요청까지 기다려야할 시간
}

type PathMatcher interface {
	Match(path string, target string) bool
}

type ApiMatchResult struct {
	IsMatch       bool
	Identifier    string
	Limit         int
	WindowSeconds int
	ExpireSeconds int
	RefillSeconds int
	Target        string
}

type QueuedRequest struct {
	Writer  http.ResponseWriter
	Request *http.Request
}
