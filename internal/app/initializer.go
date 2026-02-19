package app

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"gate-limiter/pkg/redisclient"
	"log"
)

func InitRateLimitHandler(ctx context.Context, configPath string) (*limiter.RateLimitHandler, *settings.RootRateLimiterConfig, error) {
	config, err := initConfig(configPath)
	if err != nil {
		return nil, nil, err
	}

	redisClient := NewRedisClient(&config.RedisConfig)
	keyGenerator := NewKeyGenerator(config.RateLimiter)
	if keyGenerator == nil {
		return nil, nil, fmt.Errorf("failed to initialize key generator: unsupported identity key: %q", config.RateLimiter.Identity.Key)
	}

	responder := limiter.NewHttpLimitResponder(redisClient, keyGenerator, config.RateLimiter)
	proxy := limiter.NewDefaultProxyHandler()

	rl := initRateLimiter(&config.RateLimiter, keyGenerator, &redisClient, ctx)

	return limiter.NewRateLimitHandler(rl, proxy, responder, config.RateLimiter, redisClient), config, nil
}

func initConfig(configPath string) (*settings.RootRateLimiterConfig, error) {
	if configPath == "" {
		configPath = "config.yml"
	}

	rlc, err := settings.LoadRateLimitConfig(configPath)
	if err != nil {
		log.Printf("error occur while loading rate limiter config: %v", err)
		return nil, err
	}
	return rlc, nil
}

func initRateLimiter(
	config *settings.RateLimiterConfig,
	keyGenerator *util.IpKeyGenerator,
	redisClient *types.RedisClient,
	ctx context.Context,
) types.RateLimiter {
	var rl types.RateLimiter
	log.Printf("selected strategy: [%s]\n", config.Strategy)
	switch config.Strategy {
	case "token_bucket":
		rl = strategy.NewTokenBucketLimiter(keyGenerator, *redisClient, *config)
	case "leaky_bucket":
		leakyBucketManager := strategy.NewLeakyBucketManager(ctx, config.Apis)
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
