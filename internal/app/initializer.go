package app

import (
	"errors"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/pkg/redisclient"
	"log"
)

func InitializeRateHandler() *limiter.RateLimitHandler {

	// Load config.yml
	rlc, err := config_ratelimiter.LoadRateLimitConfig("config.yml")
	if err != nil {
		log.Println("error occur while loading config.yml", err)
	}
	config := rlc.RateLimiter

	redisClient := initRedisClient(config)
	keyGenerator, err := initKeyGenerator(config)
	if err != nil {
		log.Fatalln("init key_generator fail")
	}

	responder := limiter.NewHttpLimitResponder(nil, redisClient, keyGenerator, config)
	proxy := limiter.NewDefaultProxyHandler()
	matcher := limiter.NewHttpRateLimitMatcher(keyGenerator, redisClient, config)

	return limiter.NewRateLimitHandler(matcher, proxy, responder, config)
}

func initKeyGenerator(config config_ratelimiter.RateLimiterConfig) (*limiter.IpKeyGenerator, error) {
	identity := config.Identity
	if identity.Key == "ipv4" {
		return limiter.NewIpKeyGenerator(), nil
	}
	return nil, errors.New("wrong identity key value")
}

func initRedisClient(config config_ratelimiter.RateLimiterConfig) redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
