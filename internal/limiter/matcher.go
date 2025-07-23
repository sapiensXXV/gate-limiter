package limiter

import (
	"fmt"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/pkg/redisclient"
	"log"
	"regexp"
	"time"
)

//var commentPathPattern = regexp.MustCompile(`^/api/item/[\w-]+/comment$`)

type RateLimitMatcher interface {
	IsTarget(method, url string) bool
	IsAllowed(ip string) (bool, int) // (허용 여부, 남은 허용 요청량)
}

type PathMatcher interface {
	Match(path string, target string) bool
}

type HttpRateLimitMatcher struct {
	KeyGenerator KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
} // HTTP IP와 url Path를 기반으로 처리율 제한을 검사하는 구현체

var _ RateLimitMatcher = (*HttpRateLimitMatcher)(nil)

const (
	regex = "regex"
	plain = "plain"
)

func NewHttpRateLimitMatcher(
	keyGenerator KeyGenerator,
	redisClient redisclient.RedisClient,
	config config_ratelimiter.RateLimiterConfig,
) *HttpRateLimitMatcher {
	h := &HttpRateLimitMatcher{}
	h.KeyGenerator = keyGenerator
	h.RedisClient = redisClient
	h.Config = config
	return h
}

func (rc *HttpRateLimitMatcher) IsTarget(requestMethod, requestPath string) bool {
	// 경로와 HTTP 메서드가 둘다 일치해야 제한 대상으로 판명
	apis := rc.Config.Apis
	for _, api := range apis {
		pathExpression := api.Path.Expression
		targetPath := api.Path.Value
		var result bool
		// 경로 표현 방식에 따라 경로 매칭 방식 결정
		if pathExpression == regex {
			result = rc.matchRegexPath(requestPath, targetPath)
		} else if pathExpression == plain {
			result = rc.matchPlainPath(requestPath, targetPath)
		}
		if result && requestMethod == api.Method {
			return true
		}
	}
	return false
}

func (rc *HttpRateLimitMatcher) IsAllowed(ip string) (bool, int) {
	fmt.Printf("ip_adrress: [%s]를 검사합니다.\n", ip)
	key := rc.KeyGenerator.Make(ip, rc.Config.Apis[0].Key)

	var err error
	now := time.Now()

	err = rc.RedisClient.RemoveOldEntries(key, now.Add(-1*time.Minute))
	if err != nil {
		log.Println("error while removing old entries:", err)
	}
	err = rc.RedisClient.AddToSortedSet(key, now.String(), now)
	if err != nil {
		log.Println("error while adding to sorted set:", err)
	}
	size := rc.RedisClient.GetZSetSize(key)
	if size > rc.Config.Apis[0].Limit {
		return false, 0
	}

	return true, rc.Config.Apis[0].Limit - size
}

func (rc *HttpRateLimitMatcher) matchPlainPath(requestPath string, target string) bool {
	return requestPath == target
}

func (rc *HttpRateLimitMatcher) matchRegexPath(requestPath string, target string) bool {
	r, err := regexp.Compile(target)
	if err != nil {
		log.Println("error while compile regex:", err)
	}
	return r.MatchString(requestPath)
}
