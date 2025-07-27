package app

import (
	"errors"
	"fmt"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/limiterutil"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/pkg/redisclient"
	"log"
	"net/http"
)

func InitRateLimitHandler(
	w http.ResponseWriter,

) (*limiter.RateLimitHandler, error) {

	// Load config.yml
	rlc, err := config_ratelimiter.LoadRateLimitConfig("config.yml")
	if err != nil {
		return nil, fmt.Errorf("error occur while loading config.yml: %w", err)
	}
	config := rlc.RateLimiter

	redisClient := NewRedisClient()
	keyGenerator, err := NewKeyGenerator(config)
	if err != nil {
		log.Fatalln("init key_generator fail")
	}

	responder := limiter.NewHttpLimitResponder(nil, redisClient, keyGenerator, config)
	proxy := limiter.NewDefaultProxyHandler()

	//rl := strategy.Strategies[config.Strategy](keyGenerator, redisClient, config)

	var rl strategy.RateLimiter
	switch config.Strategy {
	case "token_bucket":
		rl = strategy.NewTokenBucketLimiter(keyGenerator, redisClient, config)
	case "leaky_bucket":
		leakyBucketManager := strategy.NewLeakyBucketManager(proxy)
		rl = strategy.NewLeakyBucketLimiter(keyGenerator, redisClient, config, leakyBucketManager, nil)
	case "fixed_window_counter":
		rl = strategy.NewFixedWindowCounterLimiter(keyGenerator, redisClient, config)
	case "sliding_window_log":
		rl = strategy.NewSlidingWindowLogLimiter(keyGenerator, redisClient, config)
	case "sliding_window_counter":
	default:
	}

	return limiter.NewRateLimitHandler(rl, proxy, responder, config), nil
}

func NewKeyGenerator(config config_ratelimiter.RateLimiterConfig) (*limiterutil.IpKeyGenerator, error) {
	identity := config.Identity
	if identity.Key == "ipv4" {
		return limiterutil.NewIpKeyGenerator(), nil
	}
	return nil, errors.New("wrong identity key value")
}

func NewRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
