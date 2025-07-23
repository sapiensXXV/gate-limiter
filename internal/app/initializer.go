package app

import (
	"errors"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
	"log"
)

func InitializeRateHandler(config config_ratelimiter.RateLimiterConfig) *limiter.RateLimitHandler {

	// config의 내용에 따라서 주입해줄 구조체를 여기서 결정해야한다.

	redisClient := initRedisClient()
	keyGenerator, err := initKeyGenerator(config)
	if err != nil {
		log.Fatalln("init key_generator fail")
	}

	responder := limiter.NewHttpLimitResponder(nil, redisClient, keyGenerator, config)
	proxy := limiter.NewDefaultProxyHandler()
	matcher := limiter.NewHttpRateLimitMatcher(keyGenerator, redisClient)

	return limiter.NewRateLimitHandler(matcher, proxy, responder, config)
}

func initKeyGenerator(config config_ratelimiter.RateLimiterConfig) (*limiter.IpKeyGenerator, error) {
	identity := config.Identity
	if identity.Key == "ipv4" {
		return limiter.NewIpKeyGenerator(), nil
	}
	return nil, errors.New("wrong identity key value")
}

func initRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
