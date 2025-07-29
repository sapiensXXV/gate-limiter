package types

import "net/http"

type RateLimiter interface {
	IsTarget(method, requestPath string) (bool, *ApiMatchResult)
	IsAllowed(ip string, api *ApiMatchResult, queuedRequest *QueuedRequest) (bool, int)
}

type PathMatcher interface {
	Match(path string, target string) bool
}

type ApiMatchResult struct {
	Identifier    string
	Limit         int
	WindowSeconds int
	ExpireSeconds int
	RefillSeconds int
	BucketSize    int
	Target        string
}

type LeakyBucket struct {
	Queue      chan QueuedRequest
	BucketSize int
}

type QueuedRequest struct {
	Writer  http.ResponseWriter
	Request *http.Request
}
