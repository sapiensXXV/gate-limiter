package app

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"gate-limiter/pkg/redisclient"
	"log"
)

func InitRateLimitHandler() (*limiter.RateLimitHandler, error) {
	var config *settings.RootRateLimiterConfig
	config = initConfig()

	redisClient := NewRedisClient(&config.RedisConfig)
	keyGenerator := NewKeyGenerator(config.RateLimiter)

	responder := limiter.NewHttpLimitResponder(redisClient, keyGenerator, config.RateLimiter)
	proxy := limiter.NewDefaultProxyHandler()

	rl := initRateLimiter(&config.RateLimiter, keyGenerator, &redisClient, proxy)

	return limiter.NewRateLimitHandler(rl, proxy, responder, config.RateLimiter), nil
}

func initConfig() *settings.RootRateLimiterConfig {
	rlc, err := settings.LoadRateLimitConfig("config.yml") // Load config.yml
	if err != nil {
		log.Printf("error occur while loading config.yml: %v\n", err)
	}
	return rlc
}

func initRateLimiter(
	config *settings.RateLimiterConfig,
	keyGenerator *util.IpKeyGenerator,
	redisClient *types.RedisClient,
	proxy *limiter.DefaultProxyHandler,
) types.RateLimiter {
	var rl types.RateLimiter
	log.Printf("selected strategy: [%s]\n", config.Strategy)
	switch config.Strategy {
	case "token_bucket":
		rl = strategy.NewTokenBucketLimiter(keyGenerator, *redisClient, *config)
	case "leaky_bucket":
		leakyBucketManager := strategy.NewLeakyBucketManager(proxy, config.Apis)
		rl = strategy.NewLeakyBucketLimiter(*config, leakyBucketManager)
	case "fixed_window_counter":
		rl = strategy.NewFixedWindowCounterLimiter(keyGenerator, *redisClient, *config)
	case "sliding_window_log":
		rl = strategy.NewSlidingWindowLogLimiter(keyGenerator, *redisClient, *config)
	case "sliding_window_counter":
		rl = strategy.NewSlidingWindowCounterLimiter(keyGenerator, *redisClient, *config)
	default:
	}
	return rl
}

func NewKeyGenerator(config settings.RateLimiterConfig) *util.IpKeyGenerator {
	identity := config.Identity
	if identity.Key == "ipv4" {
		return util.NewIpKeyGenerator(config)
	}
	log.Printf("[ERROR] Wrong identity key value")
	return nil
}

func NewRedisClient(config *settings.RedisClientConfig) types.RedisClient {
	return redisclient.NewDefaultRedisClient(config)
}
