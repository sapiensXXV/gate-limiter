package strategy

type RateLimiter interface {
	IsTarget(method, requestPath string) (bool, *ApiMatchResult)
	IsAllowed(ip string, api *ApiMatchResult) (bool, int)
}

type PathMatcher interface {
	Match(path string, target string) bool
}

// ApiMatchResult limiter에서 target 매칭 이후 필요한 정보를 반환할 목적으로 만들어진 구조체
// 일종의 DTO 역할을 한다.
type ApiMatchResult struct {
	Identifier    string
	Limit         int
	WindowSeconds int
	ExpireSeconds int
	RefillSeconds int
	BucketSize    int
	Target        string
}
