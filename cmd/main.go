package main

import (
	"errors"
	config_ratelimiter "gate-limiter/config/settings"
	"gate-limiter/internal/app"
	"log"
	"net/http"
	"os"
)

func main() {
	configPath := os.Getenv("GATE_LIMITER_CONFIG")
	if configPath == "" {
		configPath = "config.yml"
	}

	_, err := config_ratelimiter.LoadRateLimitConfig(configPath)
	if err != nil {
		log.Fatal("Error loading config.yml file")
	}

	// redis-client initialization
	//redisclient.InitRedis()

	// handler
	limitHandler, err := app.InitRateLimitHandler() // 초기화가 이루어지는 시점
	if err != nil {
		log.Fatal("Error initializing rate limiter handler", err)
	}
	http.Handle("/", limitHandler)
	err = http.ListenAndServe(":8081", limitHandler) // 사용자의 요청을 받기 시작하는 지점

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed\n")
	} else if err != nil {
		log.Println("error starting server", err)
		os.Exit(1)
	}
}
