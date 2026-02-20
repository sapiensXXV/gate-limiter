package app

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/limiter/util"
	"log/slog"

	storeredis "gate-limiter/internal/limiter/store/redis"

	goredis "github.com/redis/go-redis/v9"
)

func InitRateLimitHandler(ctx context.Context, configPath string) (*limiter.RateLimitHandler, *settings.RootRateLimiterConfig, error) {
	config, err := initConfig(configPath)
	if err != nil {
		return nil, nil, err
	}

	handler, _, err := InitRateLimitHandlerWithConfig(ctx, config)
	if err != nil {
		return nil, nil, err
	}

	return handler, config, nil
}

// InitRateLimitHandlerWithConfig initializes a RateLimitHandler with a pre-loaded config.
// Returns the handler and a Redis ping function for health checks.
func InitRateLimitHandlerWithConfig(ctx context.Context, config *settings.RootRateLimiterConfig) (*limiter.RateLimitHandler, func(context.Context) error, error) {
	redisClient, err := storeredis.NewClient(&config.RedisConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	slog.Info("redis client connected")

	redisPing := func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	}

	keyGenerator := NewKeyGenerator(config.RateLimiter)
	if keyGenerator == nil {
		return nil, nil, fmt.Errorf("failed to initialize key generator: unsupported identity key: %q", config.RateLimiter.Identity.Key)
	}

	identifier := newClientIdentifier(config.RateLimiter)
	if identifier == nil {
		return nil, nil, fmt.Errorf("failed to initialize client identifier: unsupported identity key: %q", config.RateLimiter.Identity.Key)
	}

	counterStore := storeredis.NewCounterStore(redisClient)
	responder := limiter.NewHttpLimitResponder(config.RateLimiter)
	proxy := limiter.NewDefaultProxyHandler()

	rl := initRateLimiter(&config.RateLimiter, keyGenerator, redisClient, ctx)

	return limiter.NewRateLimitHandler(rl, proxy, responder, config.RateLimiter, counterStore, identifier), redisPing, nil
}

func initConfig(configPath string) (*settings.RootRateLimiterConfig, error) {
	if configPath == "" {
		configPath = "config.yml"
	}

	rlc, err := settings.LoadRateLimitConfig(configPath)
	if err != nil {
		slog.Error("config loading error", "error", err)
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
	slog.Info("strategy selected", "strategy", config.Strategy)
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
	if identity.Key == "ipv4" || identity.Key == "cookie" {
		return util.NewIpKeyGenerator(config)
	}
	slog.Error("wrong identity key value")
	return nil
}

func newClientIdentifier(config settings.RateLimiterConfig) limiter.ClientIdentifier {
	switch config.Identity.Key {
	case "ipv4":
		return &limiter.IpIdentifier{Header: config.Identity.Header}
	case "cookie":
		return limiter.NewCookieIdentifier()
	default:
		return nil
	}
}
