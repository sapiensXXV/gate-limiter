package strategy

type RateLimiter interface {
	IsTarget(method, requestPath string) (bool, *HttpMatchResult)
	IsAllowed(ip string, api *HttpMatchResult) (bool, int)
}

type PathMatcher interface {
	Match(path string, target string) bool
}

// HttpMatchResult limiter에서 target 매칭 이후 필요한 정보를 반환할 목적으로 만들어진 구조체
// 일종의 DTO 역할을 한다.
type HttpMatchResult struct {
	Key           string
	Limit         int
	WindowSeconds int
	ExpireSeconds int
	RefillSeconds int
	Target        string
}
