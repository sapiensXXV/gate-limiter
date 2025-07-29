package types

import "net/http"

type RateLimiter interface {
	IsTarget(method, requestPath string) *ApiMatchResult
	IsAllowed(ip string, api *ApiMatchResult, queuedRequest *QueuedRequest) (bool, int)
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
	BucketSize    int
	Target        string
}

type QueuedRequest struct {
	Writer  http.ResponseWriter
	Request *http.Request
}
