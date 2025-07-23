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
	rateLimiterConfig, err := config_ratelimiter.LoadRateLimitConfig("config.yml")
	if err != nil {
		log.Fatal("Error loading config.yml file")
	}
	log.Println(rateLimiterConfig)

	// redis initialization
	redisclient.InitRedis()

	// handler
	http.Handle("/", app.InitializeRateHandler())
	err = http.ListenAndServe(":8081", nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed\n")
	} else if err != nil {
		log.Println("error starting server", err)
		os.Exit(1)
	}
}
