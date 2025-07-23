package limiter

import (
	"fmt"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/pkg/redisclient"
	"log"
	"regexp"
	"time"
)

var commentPathPattern = regexp.MustCompile(`^/api/item/[\w-]+/comment$`)

type RateLimitMatcher interface {
	IsTarget(method, url string) bool
	IsAllowed(ip string) (bool, int) // (허용 여부, 남은 허용 요청량)
}

type HttpRateLimitMatcher struct {
	KeyGenerator KeyGenerator
	RedisClient  redisclient.RedisClient
	Config       config_ratelimiter.RateLimiterConfig
} // HTTP IP와 url Path를 기반으로 처리율 제한을 검사하는 구현체

var _ RateLimitMatcher = (*HttpRateLimitMatcher)(nil)

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
	log.Printf("[%s] url_path:[%s] 를 검사합니다.", requestMethod, requestPath)
	// TODO URL 패턴 매칭 고민
	apis := rc.Config.Apis
	for _, api := range apis {
		targetPath := api.Path
		targetMethod := api.Method
		if requestMethod == targetMethod && requestPath == targetPath {
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
