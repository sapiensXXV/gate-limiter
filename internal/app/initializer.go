package app

import (
	"errors"
	"fmt"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/limiter"
	"gate-limiter/internal/limiter/strategy"
	"gate-limiter/pkg/redisclient"
	"log"
)

func InitRateLimitHandler() (*limiter.RateLimitHandler, error) {

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

	var rl strategy.RateLimiter
	if config.Strategy == "token_bucket" {

	} else if config.Strategy == "leaky_bucket" {

	} else if config.Strategy == "fixed_window_counter" {

	} else if config.Strategy == "sliding_window_log" {

	} else if config.Strategy == "sliding_window_counter" {
		rl = NewSlidingWindowCounterLimiter(keyGenerator, redisClient, config)
	}

	return limiter.NewRateLimitHandler(rl, proxy, responder, config), nil
}

func NewKeyGenerator(config config_ratelimiter.RateLimiterConfig) (*limiter.IpKeyGenerator, error) {
	identity := config.Identity
	if identity.Key == "ipv4" {
		return limiter.NewIpKeyGenerator(), nil
	}
	return nil, errors.New("wrong identity key value")
}

func NewRedisClient() redisclient.RedisClient {
	return redisclient.NewDefaultRedisClient()
}
