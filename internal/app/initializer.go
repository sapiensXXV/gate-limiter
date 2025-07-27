package app

import (
	"gate-limiter/config/limiterconfig"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/pkg/redisclient"
	"log"
)

func InitRateLimitHandler() (*limiter.RateLimitHandler, error) {
	var config limiterconfig.RateLimiterConfig
	config = initConfig()

	redisClient := NewRedisClient()
	keyGenerator := NewKeyGenerator(config)

	responder := limiter.NewHttpLimitResponder(nil, redisClient, keyGenerator, config)
	proxy := limiter.NewDefaultProxyHandler()

	rl := initRateLimiter(&config, keyGenerator, &redisClient, proxy)

	return limiter.NewRateLimitHandler(rl, proxy, responder, config), nil
}

func initConfig() limiterconfig.RateLimiterConfig {
	rlc, err := limiterconfig.LoadRateLimitConfig("config.yml") // Load config.yml
	if err != nil {
		log.Printf("error occur while loading config.yml: %v\n", err)
	}
	return rlc.RateLimiter
}

func initRateLimiter(
	config *limiterconfig.RateLimiterConfig,
	keyGenerator *limiterutil.IpKeyGenerator,
	redisClient *redisclient.RedisClient,
	proxy *limiter.DefaultProxyHandler,
) strategy.RateLimiter {
	var rl strategy.RateLimiter
	switch config.Strategy {
	case "token_bucket":
		rl = strategy.NewTokenBucketLimiter(keyGenerator, *redisClient, *config)
	case "leaky_bucket":
		leakyBucketManager := strategy.NewLeakyBucketManager(proxy)
		rl = strategy.NewLeakyBucketLimiter(keyGenerator, *redisClient, *config, leakyBucketManager)
	case "fixed_window_counter":
		rl = strategy.NewFixedWindowCounterLimiter(keyGenerator, *redisClient, *config)
	case "sliding_window_log":
		rl = strategy.NewSlidingWindowLogLimiter(keyGenerator, *redisClient, *config)
	case "sliding_window_counter":
	default:
	}
	return rl
}

func NewKeyGenerator(config limiterconfig.RateLimiterConfig) *limiterutil.IpKeyGenerator {
	identity := config.Identity
	if identity.Key == "ipv4" {
		return limiterutil.NewIpKeyGenerator()
	}
	log.Printf("[ERROR] Wrong identity key value")
	return nil
}

func NewRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
