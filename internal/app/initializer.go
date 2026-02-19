package app

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log"

	storeredis "gate-limiter/internal/limiter/store/redis"
	goredis "github.com/redis/go-redis/v9"
)

func InitRateLimitHandler(ctx context.Context, configPath string) (*limiter.RateLimitHandler, *settings.RootRateLimiterConfig, error) {
	config, err := initConfig(configPath)
	if err != nil {
		return nil, nil, err
	}

	redisClient, err := storeredis.NewClient(&config.RedisConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	log.Println("redis client connection success")

	keyGenerator := NewKeyGenerator(config.RateLimiter)
	if keyGenerator == nil {
		return nil, nil, fmt.Errorf("failed to initialize key generator: unsupported identity key: %q", config.RateLimiter.Identity.Key)
	}

	counterStore := storeredis.NewCounterStore(redisClient)
	responder := limiter.NewHttpLimitResponder(config.RateLimiter)
	proxy := limiter.NewDefaultProxyHandler()

	rl := initRateLimiter(&config.RateLimiter, keyGenerator, redisClient, ctx)

	return limiter.NewRateLimitHandler(rl, proxy, responder, config.RateLimiter, counterStore), config, nil
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
	redisClient *goredis.Client,
	ctx context.Context,
) types.RateLimiter {
	var rl types.RateLimiter
	log.Printf("selected strategy: [%s]\n", config.Strategy)
	switch config.Strategy {
	case "token_bucket":
		s := storeredis.NewTokenBucketStore(redisClient)
		rl = strategy.NewTokenBucketLimiter(keyGenerator, s, *config)
	case "leaky_bucket":
		leakyBucketManager := strategy.NewLeakyBucketManager(ctx, config.Apis)
		rl = strategy.NewLeakyBucketLimiter(*config, leakyBucketManager)
	case "fixed_window_counter":
		s := storeredis.NewCounterStore(redisClient)
		rl = strategy.NewFixedWindowCounterLimiter(keyGenerator, s, *config)
	case "sliding_window_log":
		s := storeredis.NewSlidingWindowLogStore(redisClient)
		rl = strategy.NewSlidingWindowLogLimiter(keyGenerator, s, *config)
	case "sliding_window_counter":
		s := storeredis.NewSlidingWindowCounterStore(redisClient)
		rl = strategy.NewSlidingWindowCounterLimiter(keyGenerator, s, *config)
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
