package main

import (
	"errors"
	config_ratelimiter "gate-limiter/config/ratelimiter"
	"gate-limiter/internal/app"
	"gate-limiter/pkg/redisclient"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {

	// application init
	// environment variable initialization
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// read config.yml/config.yaml file
	rlc, err := config_ratelimiter.LoadRateLimitConfig("config.yml")
	if err != nil {
		log.Fatal("Error loading config.yml file")
	}
	log.Println(rlc.RateLimiter.Identity)
	log.Println(rlc.RateLimiter.Client)
	log.Println(rlc.RateLimiter.Apis)

	// redis initialization
	redisclient.InitRedis()

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
